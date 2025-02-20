package postgresDB

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"time"
	"urleater/dto"
)

const linkExpireIn = 90 * 24 * time.Hour

type Storage struct {
	pgxPool      *pgxpool.Pool
	queryBuilder squirrel.StatementBuilderType
}

func NewStorage(pgxPool *pgxpool.Pool) *Storage {
	return &Storage{
		pgxPool:      pgxPool,
		queryBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (s *Storage) CreateUser(ctx context.Context, email string, password string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	query, args, err := s.queryBuilder.
		Insert("users").
		Columns("email", "password_hash", "created_at").
		Values(email, passwordHash, time.Now().UTC().Format(time.RFC3339)).
		ToSql()

	if err != nil {
		return fmt.Errorf("CreateUser query error | %w", err)
	}

	_, err = s.pgxPool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("CreateUser query error | %w", err)
	}

	return nil

}

func (s *Storage) ChangePassword(ctx context.Context, email string, password string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query, args, err := s.queryBuilder.
		Update("users").
		Set("password_hash", passwordHash).
		Where(squirrel.Eq{"email": email}).
		ToSql()

	if err != nil {
		return fmt.Errorf("ChangePassword query error | %w", err)
	}

	_, err = s.pgxPool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("ChangePassword query error | %w", err)
	}

	return nil

}

func (s *Storage) GetUser(ctx context.Context, email string) (*dto.User, error) {
	var user dto.User

	query, args, err := s.queryBuilder.
		Select(
			"email",
			"password_hash",
			"urls_left",
		).
		From("users").
		Where(squirrel.Eq{"email": email}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("GetUser query error | %w", err)
	}

	err = s.pgxPool.QueryRow(ctx, query, args...).Scan(&user.Email, &user.PasswordHash, &user.UrlsLeft)
	if err != nil {
		return &dto.User{}, fmt.Errorf("GetUser query error | %w", err)
	}
	return &user, nil
}

func (s *Storage) VerifyUserPassword(ctx context.Context, email string, password string) error {
	user, err := s.GetUser(ctx, email)

	if err != nil {
		return fmt.Errorf("VerifyUserPassword query error | %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return fmt.Errorf("VerifyUserPassword query error | %w", err)
	}

	return nil
}

func (s *Storage) UpdateUserLinks(ctx context.Context, email string, newUrlsNumber int) (*dto.User, error) {
	var user dto.User

	query, args, err := s.queryBuilder.
		Update("users").
		Set("urls_left", newUrlsNumber).
		Where(squirrel.Eq{"email": email}).
		Suffix("RETURNING users.email, users.urls_left").
		ToSql()

	fmt.Println(query, args)

	if err != nil {
		return nil, fmt.Errorf("UpdateUserLinks query error | %w", err)
	}

	err = s.pgxPool.QueryRow(ctx, query, args...).Scan(
		&user.Email,
		&user.UrlsLeft,
	)

	if err != nil {
		return &dto.User{}, fmt.Errorf("UpdateUserLinks query error | %w", err)
	}

	return &user, nil
}

func (s *Storage) CreateShortLink(ctx context.Context, shortLink string, longLink string, userEmail string) (*dto.Link, error) {
	var link dto.Link
	expiresAt := time.Now().UTC().Add(linkExpireIn)

	query, args, err := s.queryBuilder.Insert("urls").
		Columns("short_url", "long_url", "created_at", "user_email", "expires_at", "times_visited").
		Values(shortLink, longLink, time.Now().UTC().Format(time.RFC3339), userEmail, expiresAt.Format(time.RFC3339), 0).
		Suffix("RETURNING short_url, long_url, expires_at").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("CreateShortLink query error | %w", err)
	}

	err = s.pgxPool.QueryRow(ctx, query, args...).Scan(
		&link.ShortUrl,
		&link.LongUrl,
		&link.ExpiresAt,
	)

	if err != nil {
		return nil, fmt.Errorf("CreateShortLink query error | %w", err)
	}

	return &link, nil
}

func (s *Storage) GetShortLink(ctx context.Context, shortLink string) (*dto.Link, error) {
	var link dto.Link

	query, args, err := s.queryBuilder.
		Select(
			"short_url",
			"long_url",
			"expires_at",
		).
		From("urls").
		Where(squirrel.Eq{"short_url": shortLink}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("GetShortLink query error | %w", err)
	}

	err = s.pgxPool.QueryRow(ctx, query, args...).Scan(&link.ShortUrl,
		&link.LongUrl,
		&link.ExpiresAt)

	if err != nil {
		return nil, fmt.Errorf("GetShortLink query error | %w", err)
	}
	return &link, nil
}

func (s *Storage) GetUserShortLinksWithOffsetAndLimit(ctx context.Context, email string, offset int, limit int) ([]dto.Link, error) {
	var links = make([]dto.Link, 0)

	query, args, err := s.queryBuilder.
		Select(
			"l.short_url",
			"l.long_url",
			"l.user_email",
			"l.expires_at",
		).
		From("urls l").
		Join("users u ON l.user_email = u.email").
		Where(squirrel.Eq{"l.user_email": email}).
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("GetAllUserShortLinks query error | %w", err)
	}

	rows, err := s.pgxPool.Query(ctx, query, args...)

	if err != nil {
		return nil, fmt.Errorf("GetAllUserShortLinks query error | %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var link dto.Link

		err = rows.Scan(
			&link.ShortUrl,
			&link.LongUrl,
			&link.UserEmail,
			&link.ExpiresAt,
		)

		if err != nil {
			return nil, fmt.Errorf("GetAllUserShortLinks query error | %w", err)
		}

		links = append(links, link)
	}

	return links, nil
}

func (s *Storage) GetTotalUserLinksNumber(ctx context.Context, email string) (int, error) {
	query, args, err := s.queryBuilder.
		Select(
			"COUNT(*)",
		).
		From("urls l").
		Join("users u ON l.user_email = u.email").
		Where(squirrel.Eq{"l.user_email": email}).
		ToSql()

	if err != nil {
		return 0, fmt.Errorf("GetTotalUserLinksNumber query build error | %w", err)
	}

	var totalUserLinks int

	err = s.pgxPool.QueryRow(ctx, query, args...).Scan(
		&totalUserLinks)

	if err != nil {
		return 0, fmt.Errorf("GetTotalUserLinksNumber query error | %w", err)
	}

	return totalUserLinks, nil
}

func (s *Storage) DeleteShortLink(ctx context.Context, shortLink string) error {
	query, args, err := s.queryBuilder.
		Delete("urls").
		Where(squirrel.Eq{"short_url": shortLink}).
		ToSql()

	if err != nil {
		return fmt.Errorf("DeleteShortLink query error | %w", err)
	}

	_, err = s.pgxPool.Exec(ctx, query, args...)

	if err != nil {
		return fmt.Errorf("DeleteShortLink query error | %w", err)
	}

	return nil
}

func (s *Storage) ExtendShortLink(ctx context.Context, shortLink string, expiresAt time.Time) (*dto.Link, error) {
	var link dto.Link

	query, args, err := s.queryBuilder.
		Update("urls").
		Set("expires_at", expiresAt.Add(linkExpireIn).UTC().Format(time.RFC3339)).
		Where(squirrel.Eq{"short_url": shortLink}).
		Suffix("RETURNING short_url, long_url, expires_at").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("ExtendShortLink query error | %w", err)
	}

	err = s.pgxPool.QueryRow(ctx, query, args...).Scan(&link.ShortUrl,
		&link.LongUrl,
		&link.ExpiresAt)

	if err != nil {
		return nil, fmt.Errorf("ExtendShortLink query error | %w", err)
	}

	return &link, nil
}

func (s *Storage) GetSubscriptions(ctx context.Context) ([]dto.Subscription, error) {
	var subscriptions []dto.Subscription

	query, args, err := s.queryBuilder.
		Select(
			"id",
			"name",
			"total_urls",
		).
		From("subscriptions").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("GetSubscriptions query error | %w", err)
	}

	rows, err := s.pgxPool.Query(ctx, query, args...)

	if err != nil {
		return nil, fmt.Errorf("GetSubscriptions query error | %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var sub dto.Subscription
		err = rows.Scan(
			&sub.Id,
			&sub.Name,
			&sub.TotalUrls,
		)

		if err != nil {
			return nil, fmt.Errorf("GetSubscriptions scan error | %w", err)
		}

		subscriptions = append(subscriptions, sub)

	}

	return subscriptions, nil

}

func (s *Storage) CreateSubscriptions(ctx context.Context) error {
	query, args, err := s.queryBuilder.Insert("subscriptions").
		Columns("name", "total_urls").
		Values("Bronze", 1000).
		Values("Silver", 5000).
		Values("Gold", 10000).
		ToSql()

	if err != nil {
		return fmt.Errorf("CreateSubscriptions query build error | %w", err)
	}

	_, err = s.pgxPool.Exec(ctx, query, args...)

	if err != nil {
		return fmt.Errorf("CreateShortLink query error | %w", err)
	}

	return nil
}

func (s *Storage) IncrementShortLinkTimesWatchedCount(ctx context.Context, shortLink string) error {
	query, args, err := s.queryBuilder.
		Update("urls").
		Set("times_visited", squirrel.Expr("times_visited + 1")).
		Where(squirrel.Eq{"short_url": shortLink}).
		ToSql()

	if err != nil {
		return fmt.Errorf("IncrementShortLinkTimesWatchedCount query error | %w", err)
	}

	_, err = s.pgxPool.Exec(ctx, query, args...)

	if err != nil {
		return fmt.Errorf("IncrementShortLinkTimesWatchedCount query error | %w", err)
	}

	return nil
}
