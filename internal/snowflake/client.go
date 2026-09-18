package snowflake

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"

	sf "github.com/snowflakedb/gosnowflake/v2"
)

var ident = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_$]*$`)

type Client struct {
	DB *sql.DB
}

func New() (*Client, error) {
	account := os.Getenv("SNOWFLAKE_ACCOUNT")
	user := os.Getenv("SNOWFLAKE_USER")
	tokenFile := os.Getenv("SNOWFLAKE_TOKEN_FILE")

	if account == "" {
		return nil, fmt.Errorf("SNOWFLAKE_ACCOUNT is required")
	}

	if user == "" {
		return nil, fmt.Errorf("SNOWFLAKE_USER is required")
	}

	if tokenFile == "" {
		return nil, fmt.Errorf("SNOWFLAKE_TOKEN_FILE is required")
	}

	cfg := &sf.Config{
		Account:                  account,
		User:                     user,
		Authenticator:            sf.AuthTypeWorkloadIdentityFederation,
		WorkloadIdentityProvider: "OIDC",
		TokenFilePath:            tokenFile,
	}

	connector := sf.NewConnector(
		sf.SnowflakeDriver{},
		*cfg,
	)

	dbh := sql.OpenDB(connector)

	return &Client{
		DB: dbh,
	}, nil
}

func (c *Client) Close() {
	if c != nil && c.DB != nil {
		_ = c.DB.Close()
	}
}

func Quote(s string) (string, error) {
	if !ident.MatchString(s) {
		return "", fmt.Errorf(
			"invalid Snowflake identifier %q",
			s,
		)
	}

	return `"` + s + `"`, nil
}

func Lit(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func (c *Client) Exec(ctx context.Context, q string) error {
	_, err := c.DB.ExecContext(ctx, q)
	return err
}

func (c *Client) Exists(
	ctx context.Context,
	kind string,
	name string,
) (bool, error) {

	var n int
	var q string

	v := Lit(name)

	switch strings.ToUpper(kind) {

	case "ROLE":
		q = `
			SELECT COUNT(*)
			FROM SNOWFLAKE.ACCOUNT_USAGE.ROLES
			WHERE NAME = ` + v + `
			  AND DELETED_ON IS NULL
		`

	case "USER":
		q = `
			SELECT COUNT(*)
			FROM SNOWFLAKE.ACCOUNT_USAGE.USERS
			WHERE NAME = ` + v + `
			  AND DELETED_ON IS NULL
		`

	case "DATABASE":
		q = `
			SELECT COUNT(*)
			FROM SNOWFLAKE.ACCOUNT_USAGE.DATABASES
			WHERE DATABASE_NAME = ` + v + `
			  AND DELETED_ON IS NULL
		`

	case "WAREHOUSE":
		q = `
			SELECT COUNT(*)
			FROM SNOWFLAKE.ACCOUNT_USAGE.WAREHOUSES
			WHERE NAME = ` + v + `
			  AND DELETED_ON IS NULL
		`

	default:
		return false, fmt.Errorf(
			"unsupported kind %s",
			kind,
		)
	}

	err := c.DB.QueryRowContext(ctx, q).Scan(&n)

	return n > 0, err
}

func (c *Client) EnsureRole(
	ctx context.Context,
	name string,
	comment string,
) error {

	q, err := Quote(name)
	if err != nil {
		return err
	}

	stmt := "CREATE ROLE IF NOT EXISTS " + q

	if comment != "" {
		stmt += " COMMENT = " + Lit(comment)
	}

	return c.Exec(ctx, stmt)
}

func (c *Client) EnsureUser(
	ctx context.Context,
	name string,
	role string,
	comment string,
) error {

	q, err := Quote(name)
	if err != nil {
		return err
	}

	ok, err := c.Exists(ctx, "USER", name)
	if err != nil {
		return err
	}

	if !ok {
		stmt := "CREATE USER " + q + " TYPE = SERVICE"

		if comment != "" {
			stmt += " COMMENT = " + Lit(comment)
		}

		if err := c.Exec(ctx, stmt); err != nil {
			return err
		}
	}

	if role != "" {
		r, err := Quote(role)
		if err != nil {
			return err
		}

		roleName := strings.Trim(r, `"`)

		return c.Exec(
			ctx,
			"ALTER USER "+q+
				" SET DEFAULT_ROLE = "+Lit(roleName),
		)
	}

	return nil
}

func (c *Client) EnsureDatabase(
	ctx context.Context,
	name string,
	comment string,
) error {

	q, err := Quote(name)
	if err != nil {
		return err
	}

	stmt := "CREATE DATABASE IF NOT EXISTS " + q

	if comment != "" {
		stmt += " COMMENT = " + Lit(comment)
	}

	return c.Exec(ctx, stmt)
}

func (c *Client) EnsureSchema(
	ctx context.Context,
	database string,
	name string,
	comment string,
) error {

	d, err := Quote(database)
	if err != nil {
		return err
	}

	n, err := Quote(name)
	if err != nil {
		return err
	}

	stmt := "CREATE SCHEMA IF NOT EXISTS " +
		d + "." + n

	if comment != "" {
		stmt += " COMMENT = " + Lit(comment)
	}

	return c.Exec(ctx, stmt)
}

func (c *Client) EnsureWarehouse(
	ctx context.Context,
	name string,
	size string,
	autosuspend int,
	comment string,
) error {

	q, err := Quote(name)
	if err != nil {
		return err
	}

	stmt := "CREATE WAREHOUSE IF NOT EXISTS " + q

	if size != "" {
		stmt += " WAREHOUSE_SIZE = " + Lit(size)
	}

	if autosuspend > 0 {
		stmt += fmt.Sprintf(
			" AUTO_SUSPEND = %d",
			autosuspend,
		)
	}

	if comment != "" {
		stmt += " COMMENT = " + Lit(comment)
	}

	return c.Exec(ctx, stmt)
}

func (c *Client) Grant(
	ctx context.Context,
	priv string,
	on string,
	obj string,
	role string,
) error {

	p, err := Quote(priv)
	if err != nil {
		return err
	}

	o, err := Quote(obj)
	if err != nil {
		return err
	}

	r, err := Quote(role)
	if err != nil {
		return err
	}

	k := strings.ToUpper(strings.TrimSpace(on))

	switch k {
	case "ROLE",
		"DATABASE",
		"SCHEMA",
		"WAREHOUSE",
		"TABLE",
		"VIEW":
	default:
		return fmt.Errorf(
			"unsupported grant object kind %q",
			on,
		)
	}

	return c.Exec(
		ctx,
		fmt.Sprintf(
			"GRANT %s ON %s %s TO ROLE %s",
			strings.Trim(p, `"`),
			k,
			o,
			r,
		),
	)
}
