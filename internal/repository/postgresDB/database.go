package postgresDB

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"time"
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
	query, args, err := s.queryBuilder.Insert("users").
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

func (s *Storage) GetUser(ctx context.Context, email string) (*User, error) {
	var user User

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
		return &User{}, fmt.Errorf("GetUser query error | %w", err)
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

func (s *Storage) UpdateUserLinks(ctx context.Context, email string, urlsDelta int) (*User, error) {
	var user User

	query, args, err := s.queryBuilder.
		Update("users").
		Set("urls_left", squirrel.Expr("IF(urls_left + ? >= 0, urls_left + ?, 0)", urlsDelta, urlsDelta)).
		Where(squirrel.Eq{"email": email}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("UpdateUserLinks query error | %w", err)
	}

	err = s.pgxPool.QueryRow(ctx, query, args...).Scan(
		&user.Email,
		&user.PasswordHash,
		&user.UrlsLeft,
	)

	if err != nil {
		return &User{}, fmt.Errorf("UpdateUserLinks query error | %w", err)
	}

	return &user, nil
}

func (s *Storage) CreateShortLink(ctx context.Context, shortLink string, longLink string, userEmail string) (*Link, error) {
	var link Link
	expiresAt := time.Now().UTC().Add(linkExpireIn)
	query, args, err := s.queryBuilder.Insert("urls").
		Columns("short_url", "long_url", "created_at", "user_email", "expires_at").
		Values(shortLink, longLink, time.Now().UTC().Format(time.RFC3339), userEmail, expiresAt.Format(time.RFC3339)).
		Suffix("RETURNING short_url, long_url, expires_at").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("CreateShortLink query error | %w", err)
	}

	err = s.pgxPool.QueryRow(ctx, query, args...).Scan(
		&link.ShortUrl,
		&link.LongUrl,
		&link.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("CreateShortLink query error | %w", err)
	}

	return &link, nil
}

func (s *Storage) GetShortLink(ctx context.Context, shortLink string) (*Link, error) {
	var link Link

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

func (s *Storage) GetUserShortLinksWithOffsetAndLimit(ctx context.Context, email string, offset int, limit int) ([]Link, error) {
	var links []Link

	query, args, err := s.queryBuilder.
		Select(
			"l.short_url",
			"l.long_url",
			"l.user_email",
			"l.expires_at",
		).
		From("urls l").
		Join("users u ON urls.user_email = users.email").
		Where(squirrel.Eq{"urls.user_email": email}).
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
		var link Link

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

func (s *Storage) DeleteShortLink(ctx context.Context, shortLink string, email string) error {
	query, args, err := s.queryBuilder.
		Delete("urls").
		Where(squirrel.Eq{"short_url": shortLink}).
		Where(squirrel.Eq{"email": email}).
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

func (s *Storage) ExtendShortLink(ctx context.Context, shortLink string, expiresAt time.Time) (*Link, error) {
	var link Link

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

func (s *Storage) GetSubscriptions(ctx context.Context) ([]Subscription, error) {
	var subscriptions []Subscription

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
		var sub Subscription
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
