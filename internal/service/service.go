package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"log"
	"math/rand"
	"net/mail"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode"
	"urleater/dto"
	kafkaProducerConsumer "urleater/internal/repository/kafka"
)

type PostgresStorage interface {
	CreateUser(ctx context.Context, email string, password string) error
	ChangePassword(ctx context.Context, email string, password string) error
	GetUser(ctx context.Context, email string) (*dto.User, error)
	CreateShortLink(ctx context.Context, shortLink string, longLink string, userID string) (*dto.Link, error)
	GetShortLink(ctx context.Context, shortLink string) (*dto.Link, error)
	DeleteShortLink(ctx context.Context, shortLink string) error
	ExtendShortLink(ctx context.Context, shortLink string, expiresAt time.Time) (*dto.Link, error)
	GetUserShortLinksWithOffsetAndLimit(ctx context.Context, email string, offset int, limit int) ([]dto.Link, error)
	UpdateUserLinks(ctx context.Context, email string, newUrlsNumber int) (*dto.User, error)
	GetSubscriptions(ctx context.Context) ([]dto.Subscription, error)
	VerifyUserPassword(ctx context.Context, email string, password string) error
	CreateSubscriptions(ctx context.Context) error
	GetTotalUserLinksNumber(ctx context.Context, email string) (int, error)
	IncrementShortLinkTimesWatchedCount(ctx context.Context, shortLink string) error
}

type RedisStorage interface {
	DeleteLongLinkByShortLink(ctx context.Context, shortLink string) error
	GetShortLinkByLongLink(ctx context.Context, shortLink string) (*dto.Link, error)
	SaveShortLinkToLongLink(ctx context.Context, link dto.Link) error
}

type Consumer interface {
	StartConsuming(ctx context.Context) error
	GetConfig() kafkaProducerConsumer.KafkaConfig
	GetWorkerChannel() chan dto.ConsumerData
}

type ElasticSearcher interface {
	SearchShortLinks(ctx context.Context, word string, limit int, offset int) ([]string, error)
	AddShortLink(ctx context.Context, link string) error
	DeleteShortLink(ctx context.Context, link string) error
}

type Producer interface {
	PublishMsg(msgType string, data map[string]string, topic string) error
	GetConfig() kafkaProducerConsumer.KafkaConfig
}

var mutex = &sync.Mutex{}

type Service struct {
	postgresStorage PostgresStorage
	redisStorage    RedisStorage
	consumers       []Consumer
	producer        Producer
	searcher        ElasticSearcher
	producerTopic   string
}

var reservedNames = []string{
	"register",
	"login",
	"logout",
	"create_link",
	"buy",
	"subscriptions",
}

func New(postgresStorage PostgresStorage, redisStorage RedisStorage, producer Producer, consumers []Consumer, searcher ElasticSearcher, producerTopic string) *Service {
	return &Service{
		postgresStorage: postgresStorage,
		redisStorage:    redisStorage,
		consumers:       consumers,
		producer:        producer,
		searcher:        searcher,
		producerTopic:   producerTopic,
	}
}

func (s *Service) LoginUser(ctx context.Context, email string, password string) error {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	if len(email) == 0 || len(password) == 0 {
		return fmt.Errorf("LoginUser: email or password is empty")
	}

	if !validateEmail(email) {
		return fmt.Errorf("LoginUser: invalid email format")
	}

	err := s.postgresStorage.VerifyUserPassword(ctx, email, password)

	switch {
	case err == nil:

	case errors.Is(err, pgx.ErrNoRows):
		return fmt.Errorf("LoginUser: invalid password")

	default:
		return fmt.Errorf("LoginUser: could not verify password %w", err)

	}

	return nil
}

func validatePassword(password string) bool {
	// Проверка длины пароля (не меньше 8 символов)
	if len(password) < 8 {
		return false
	}

	// Проверка на наличие только допустимых символов
	for _, char := range password {
		if !(unicode.IsDigit(char) || isSpecialCharacter(char)) &&
			(char < 'a' || char > 'z') && (char < 'A' || char > 'Z') {
			return false
		}
	}

	return true
}

// Функция для проверки спецсимволов
func isSpecialCharacter(char rune) bool {
	specialCharacters := "!@#$%^&*()-_=+[]{}|;:'\",.<>?/`~"
	for _, special := range specialCharacters {
		if char == special {
			return true
		}
	}
	return false
}

func validateEmail(email string) bool {
	for _, char := range email {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && !strings.Contains("!-.@#_", string(char)) {
			return false
		}
	}
	_, err := mail.ParseAddress(email)

	return err == nil
}

func (s *Service) RegisterUser(ctx context.Context, email string, password string) error {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	if len(email) == 0 || len(password) == 0 {
		return fmt.Errorf("RegisterUser: email or password is empty")
	}

	if !validatePassword(password) {
		return fmt.Errorf("RegisterUser: invalid password format")
	}

	if !validateEmail(email) {
		return fmt.Errorf("RegisterUser: invalid email format")
	}

	_, err := s.postgresStorage.GetUser(ctx, email)

	switch {
	case errors.Is(err, pgx.ErrNoRows):

	case err != nil:
		return fmt.Errorf("RegisterUser: could not get user %w", err)
	default:
		return fmt.Errorf("RegisterUser: user already exists")
	}

	err = s.postgresStorage.CreateUser(ctx, email, password)

	if err != nil {
		return fmt.Errorf("RegisterUser: could not create user %w", err)
	}

	return nil
}

func validateLinkAlias(alias string) bool {
	if len(alias) < 8 || len(alias) > 20 {
		return false
	}
	for _, char := range alias {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && !unicode.IsDigit(char) {
			return false
		}
	}

	return true
}

func IsValidUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func (s *Service) CreateShortLink(ctx context.Context, alias string, longLink string, userEmail string) (*dto.Link, error) {
	fmt.Println()
	if len(longLink) == 0 {
		return nil, fmt.Errorf("CreateShortLink: longLink is empty")
	}

	if !IsValidUrl(longLink) {
		return nil, fmt.Errorf("CreateShortLink: invalid longLink format")
	}

	var shortLink string

	if alias != "" {
		if !validateLinkAlias(alias) {
			return nil, fmt.Errorf("CreateShortLink: invalid alias: %s", alias)
		}
		shortLink = alias
	} else {
		var err error

	forLoop:
		for i := 0; i < 10; i++ { // генерируем ссылки, пока такие существуют
			shortLink = GenerateShortLink()

			_, err = s.postgresStorage.GetShortLink(ctx, shortLink)

			switch {
			case errors.Is(err, pgx.ErrNoRows):
				break forLoop

			case err != nil:
				return nil, fmt.Errorf("failed to check if short link exists: %w", err)

			case i == 9:
				return nil, fmt.Errorf("could not generate short link in 10 tries")

			}
		}
	}

	for _, val := range reservedNames {
		if val == shortLink {
			return nil, fmt.Errorf("short link %s is not available", shortLink)
		}
	}

	user, err := s.postgresStorage.GetUser(ctx, userEmail)

	if err != nil {
		return nil, fmt.Errorf("CreateShortLink: could not get user: %w", err)
	}

	if user.UrlsLeft == 0 {
		return nil, fmt.Errorf("CreateShortLink: user %s has no urls", userEmail)
	}

	_, err = s.postgresStorage.UpdateUserLinks(ctx, userEmail, user.UrlsLeft-1)

	if err != nil {
		return nil, fmt.Errorf("CreateShortLink: error while updating user links for short link %s | %w", shortLink, err)
	}

	link, err := s.postgresStorage.CreateShortLink(ctx, shortLink, longLink, userEmail)

	if err != nil {
		return nil, fmt.Errorf("CreateShortLink: error while creating a short link %s | %w", shortLink, err)
	}

	err = s.redisStorage.SaveShortLinkToLongLink(ctx, *link)

	if err != nil {
		log.Println(fmt.Errorf("CreateShortLink: error while saving short link %s | %w", shortLink, err).Error())
	}

	err = s.searcher.AddShortLink(ctx, shortLink)

	if err != nil {
		log.Println(fmt.Errorf("CreateShortLink: error while adding short link %s | %v", shortLink, err).Error())
	}

	return link, nil
}

func (s *Service) GetSubscriptions(ctx context.Context) ([]dto.Subscription, error) {
	subs, err := s.postgresStorage.GetSubscriptions(ctx)

	if err != nil {
		return nil, fmt.Errorf("GetSubscriptions: could not get subscriptions %w", err)
	}

	return subs, nil
}

func (s *Service) GetUser(ctx context.Context, email string) (*dto.User, error) {
	user, err := s.postgresStorage.GetUser(ctx, email)

	if err != nil {
		return nil, fmt.Errorf("GetUser: could not get user %w", err)
	}

	return user, nil
}

func (s *Service) GetUserShortLinksWithOffsetAndLimit(ctx context.Context, email string, offset int, limit int) ([]dto.Link, *dto.User, error) {
	user, err := s.postgresStorage.GetUser(ctx, email)

	if err != nil {
		return nil, nil, fmt.Errorf("GetUserShortLinksWithOffsetAndLimit: error while getting user %s: %w", email, err)
	}

	if limit == 0 || limit > 50 {
		limit = 50
	}

	links, err := s.postgresStorage.GetUserShortLinksWithOffsetAndLimit(ctx, email, offset, limit)

	switch {
	case err == nil:

	case errors.Is(err, pgx.ErrNoRows):

	default:
		return nil, nil, fmt.Errorf("GetUserShortLinksWithOffsetAndLimit: error while getting all user's %s shortlinks: %w", email, err)
	}

	return links, user, nil

}

func (s *Service) GetTotalUserLinks(ctx context.Context, email string) (int, error) {
	totalUserLinks, err := s.postgresStorage.GetTotalUserLinksNumber(ctx, email)

	if err != nil {
		return 0, fmt.Errorf("GetTotalUserLinks: %w", err)
	}

	return totalUserLinks, nil
}

func (s *Service) UpdateUserShortLinks(ctx context.Context, email string, deltaLinks int) (*dto.User, error) {
	user, err := s.postgresStorage.UpdateUserLinks(ctx, email, deltaLinks)

	if err != nil {
		return nil, fmt.Errorf("UpdateUserShortLinks: error while updating user's %s shortlinks: %w by %d", email, err, deltaLinks)
	}

	return user, nil
}

func (s *Service) GetShortLink(ctx context.Context, shortLink string) (*dto.Link, error) {
	link, err := s.redisStorage.GetShortLinkByLongLink(ctx, shortLink)

	if err != nil {
		// TODO log
		link, err = s.postgresStorage.GetShortLink(ctx, shortLink)

		if err != nil {
			return nil, fmt.Errorf("GetShortLink: error while getting short link %s: %w", shortLink, err)
		}

		err = s.redisStorage.SaveShortLinkToLongLink(ctx, *link)

		if err != nil {
			log.Println(fmt.Errorf("GetShortLink: error while saving short link  to redis %s: %v", shortLink, err).Error())
		}
	}

	if link.ExpiresAt.Before(time.Now()) {
		go func() {
			err := s.producer.PublishMsg("increment_link_view", map[string]string{
				"short_link": shortLink,
			}, s.producerTopic)

			if err != nil {
				log.Println(fmt.Errorf("error while publishing to %s topic: %w", s.producerTopic, err).Error())

				s.recreateProducer(ctx)
			}

		}()

		return nil, fmt.Errorf("GetShortLink: short link %s expired", shortLink)
	} else {
		go func() {
			err := s.producer.PublishMsg("increment_link_view", map[string]string{
				"short_link": shortLink,
			}, s.producerTopic)

			if err != nil {
				log.Println(fmt.Errorf("error while publishing to %s topic: %w", s.producerTopic, err).Error())

				s.recreateProducer(ctx)
			} else {
				log.Println("sent to consumer")
			}
		}()
	}

	return link, nil

}

func (s *Service) recreateProducer(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)

	var err error

innerLoop:
	for {
		select {
		case <-ctx.Done():
			// TODO close consumers
			return
		case <-ticker.C:
			s.producer, err = kafkaProducerConsumer.NewProducer(s.producer.GetConfig())

			if err == nil {
				ticker.Stop()

				break innerLoop
			}
		}
	}
}

func (s *Service) DeleteShortLink(ctx context.Context, shortLink string, email string) error {
	link, err := s.postgresStorage.GetShortLink(ctx, shortLink)

	if err != nil {
		return fmt.Errorf("DeleteShortLink: error while getting short link %s: %w", shortLink, err)
	}

	fmt.Println(link.UserEmail, email)
	if link.UserEmail != email {
		return fmt.Errorf("DeleteShortLink: short link %s does not match email %s", shortLink, email)
	}

	err = s.postgresStorage.DeleteShortLink(ctx, shortLink)

	if err != nil {
		return fmt.Errorf("DeleteShortLink: error while deleting short link %s with email %s: %w", shortLink, email, err)
	}

	err = s.redisStorage.DeleteLongLinkByShortLink(ctx, shortLink)

	if err != nil {
		log.Println(fmt.Errorf("DeleteShortLink: error while deleting short link from redis %s: %w", shortLink, err).Error())
	}

	err = s.searcher.DeleteShortLink(ctx, shortLink)

	if err != nil {
		log.Println(fmt.Errorf("DeleteShortLink: error while deleting short link from elastic %s: %w", shortLink, err).Error())
	}

	return nil
}

func (s *Service) CreateSubscriptions(ctx context.Context) error {
	err := s.postgresStorage.CreateSubscriptions(ctx)

	if err != nil {
		return fmt.Errorf("CreateSubscriptions: error while creating subscriptions %w", err)
	}

	return nil
}

const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateShortLink() string {
	mutex.Lock()
	defer mutex.Unlock()
	source := rand.NewSource(time.Now().UnixNano())

	res := make([]byte, 8)

	for i := range res {
		res[i] = letterBytes[source.Int63()%int64(len(letterBytes))]
	}

	return string(res)

}

func (s *Service) StartConsumers(ctx context.Context) {
	for i := 0; i < len(s.consumers); i++ {
		go func(i int) {
			for {
				err := s.consumers[i].StartConsuming(ctx)

				if err == nil {
					return
				}

				log.Println("error while running kafka consumer", err)

				ticker := time.NewTicker(5 * time.Second)

				log.Println("recreating kafka consumer...")
			innerLoop:
				for {
					select {
					case <-ctx.Done():
						// TODO close consumers
						return
					case <-ticker.C:
						s.consumers[i], err = kafkaProducerConsumer.NewConsumer(s.consumers[i].GetConfig(), s.consumers[i].GetWorkerChannel())

						if err == nil {
							ticker.Stop()

							log.Println("recreated kafka consumer")
							break innerLoop
						}
					}

				}
			}
		}(i)
	}
}

func (s *Service) StartConsumingWorkers(ctx context.Context, n int, workerChannel chan dto.ConsumerData) {
	for i := 0; i < n; i++ {
		go func() {
			for {
				err := s.startWorker(ctx, workerChannel)

				if err == nil {
					return
				}
			}
		}()
	}
}

func (s *Service) startWorker(ctx context.Context, workerChannel chan dto.ConsumerData) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case data := <-workerChannel:
			err := s.processData(ctx, data.TypeOfMessage, data.Data)

			if err != nil {
				log.Println(fmt.Errorf("error while processing data: %w", err))

				return fmt.Errorf("startWorker: error while processing data: %w", err)
			}
		}
	}

}

func (s *Service) processData(ctx context.Context, dataType string, data map[string]string) error {
	log.Printf("data type: %s, data: %v\n", dataType, data)
	switch dataType {
	case "delete_expired_link":
		shortLink, ok := data["short_link"]

		if !ok {
			return fmt.Errorf("processData: invalid data type for delete_expired_link")
		}

		err := s.redisStorage.DeleteLongLinkByShortLink(ctx, shortLink)

		if err != nil {
			return fmt.Errorf("processData: error while deleting short link from redis: %w", err)
		}

		err = s.postgresStorage.DeleteShortLink(ctx, shortLink)

		if err != nil {
			return fmt.Errorf("processData: error while deleting short link from postgres: %w", err)
		}

		err = s.searcher.DeleteShortLink(ctx, shortLink)

		if err != nil {
			log.Println(fmt.Errorf("DeleteShortLink: error while deleting short link from searcher %s: %w", shortLink, err).Error())
		}

		return nil
	case "increment_link_view":
		shortLink, ok := data["short_link"]

		if !ok {
			return fmt.Errorf("processData: invalid data type for delete_expired_link")
		}

		err := s.postgresStorage.IncrementShortLinkTimesWatchedCount(ctx, shortLink)

		if err != nil {
			return fmt.Errorf("processData: error while incrementing short link times: %w", err)
		}

		return nil
	default:
		return nil
	}
}

func (s *Service) GetShortLinksMatchingPattern(ctx context.Context, containsWord string, offset int) (dto.SearcherMatchResult, error) {
	if offset < 0 {
		return dto.SearcherMatchResult{}, nil
	}

	limit := 20

	shortLinks, err := s.searcher.SearchShortLinks(ctx, containsWord, limit, offset)

	if err != nil {
		return dto.SearcherMatchResult{}, fmt.Errorf("GetShortLinksMatchingPattern: error while searching short links: %w", err)
	}

	return dto.SearcherMatchResult{
		Limit:      limit,
		Offset:     offset,
		ShortLinks: shortLinks,
	}, nil
}

func (s *Service) LoginUserWithCode(ctx context.Context, email string) error {
	email = strings.TrimSpace(email)

	if len(email) == 0 {
		return fmt.Errorf("LoginUser: email  is empty")
	}

	if !validateEmail(email) {
		return fmt.Errorf("LoginUser: invalid email format")
	}

	return nil

}
