package pkg

import (
	"context"
	"errors"
	"io/ioutil"
	"path/filepath"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/syols/keeper/internal/models"
)

const ScriptPath = "scripts/query/"
const MigrationPath = "file://scripts/migrations/"

type ReadDetailFunc func(ctx context.Context, record models.Record) (models.Details, error)
type RelativePath string

type Database struct {
	Scripts      map[string]string
	connection   DatabaseConnectionCreator
	readDetails  map[string]ReadDetailFunc
	queryDetails map[string]string
}

func NewDatabase(connectionCreator DatabaseConnectionCreator) (db Database, err error) {
	db = Database{
		Scripts:    map[string]string{},
		connection: connectionCreator,
	}
	db.readDetails = map[string]ReadDetailFunc{
		models.TextType:  db.textDetails,
		models.BlobType:  db.blobDetails,
		models.CardType:  db.cardDetails,
		models.LoginType: db.loginDetails,
	}

	db.queryDetails = map[string]string{
		models.TextType:  "details/create_text.sql",
		models.BlobType:  "details/create_blob.sql",
		models.CardType:  "details/create_card.sql",
		models.LoginType: "details/create_login.sql",
	}
	err = connectionCreator.Migrate()
	return
}

func (d *Database) Register(ctx context.Context, user *models.User) error {
	rows, err := d.execute(ctx, "create_user.sql", user)
	if err != nil {
		return err
	}

	if err := rows.Err(); err != nil {
		return err
	}
	return err
}

func (d *Database) Login(ctx context.Context, user *models.User) (*models.User, error) {
	rows, err := d.execute(ctx, "user.sql", user)
	if err != nil {
		return nil, err
	}
	return bindUser(rows)
}

func (d *Database) UserRecords(ctx context.Context, username string) ([]models.Record, error) {
	user, err := d.user(ctx, username)
	if err != nil {
		return nil, err
	}

	rows, err := d.execute(ctx, "records.sql", models.Record{UserID: user.ID})
	if err != nil {
		return nil, err
	}

	var values []models.Record
	for rows.Next() {
		var value models.Record
		if err := rows.StructScan(&value); err != nil {
			return nil, err
		}

		detailFunc, isOk := d.readDetails[value.DetailType]
		if !isOk {
			return nil, errors.New("incorrect detail type")
		}

		privateData, err := detailFunc(ctx, value)
		if err != nil {
			return nil, err
		}

		value.PrivateData = privateData
		values = append(values, value)
	}
	return values, nil
}

func (d *Database) FindRecords(ctx context.Context, username string) ([]models.Record, error) {
	user, err := d.user(ctx, username)
	if err != nil {
		return nil, err
	}

	rows, err := d.execute(ctx, "records.sql", models.Record{UserID: user.ID})
	if err != nil {
		return nil, err
	}

	var values []models.Record
	for rows.Next() {
		var value models.Record
		if err := rows.StructScan(&value); err != nil {
			return nil, err
		}

		detailFunc, isOk := d.readDetails[value.DetailType]
		if !isOk {
			return nil, errors.New("incorrect detail type")
		}

		privateData, err := detailFunc(ctx, value)
		if err != nil {
			return nil, err
		}

		value.PrivateData = privateData
		values = append(values, value)
	}
	return values, nil
}

func (d *Database) SaveRecord(ctx context.Context, username string, record models.Record) error {
	user, err := d.user(ctx, username)
	if err != nil {
		return err
	}
	record.UserID = user.ID

	rows, err := d.execute(ctx, "create_record.sql", record)
	if err != nil {
		return err
	}

	var value models.Record
	if rows.Next() {
		if err := rows.StructScan(&value); err != nil {
			return err
		}
	}

	details := record.PrivateData.SetRecordId(value.ID)
	query, isOk := d.queryDetails[value.DetailType]
	if !isOk {
		return errors.New("incorrect record detail type")
	}

	_, err = d.execute(ctx, query, details)
	return err
}

func (d *Database) Truncate(ctx context.Context) error {
	_, err := d.execute(ctx, "truncate.sql", models.Record{})
	return err
}

func (d *Database) textDetails(ctx context.Context, record models.Record) (models.Details, error) {
	rows, err := d.execute(ctx, "details/text.sql", record)
	if err != nil {
		return nil, err
	}

	var value models.TextDetails
	if rows.Next() {
		if err := rows.StructScan(&value); err != nil {
			return nil, err
		}
	}
	return value, nil
}

func (d *Database) blobDetails(ctx context.Context, record models.Record) (models.Details, error) {
	rows, err := d.execute(ctx, "details/blob.sql", record)
	if err != nil {
		return nil, err
	}

	var value models.BlobDetails
	if rows.Next() {
		if err := rows.StructScan(&value); err != nil {
			return nil, err
		}
	}
	return value, nil
}

func (d *Database) cardDetails(ctx context.Context, record models.Record) (models.Details, error) {
	rows, err := d.execute(ctx, "details/card.sql", record)
	if err != nil {
		return nil, err
	}

	var value models.CardDetails
	if rows.Next() {
		if err := rows.StructScan(&value); err != nil {
			return nil, err
		}
	}
	return value, nil
}

func (d *Database) loginDetails(ctx context.Context, record models.Record) (models.Details, error) {
	rows, err := d.execute(ctx, "details/login.sql", record)
	if err != nil {
		return nil, err
	}

	var value models.LoginDetails
	if rows.Next() {
		if err := rows.StructScan(&value); err != nil {
			return nil, err
		}
	}
	return value, nil
}

func (d *Database) user(ctx context.Context, username string) (*models.User, error) {
	rows, err := d.execute(ctx, "user.sql", models.User{Username: username})
	if err != nil {
		return nil, err
	}
	return bindUser(rows)
}

func (d *Database) execute(ctx context.Context, filename string, model interface{}) (*sqlx.Rows, error) {
	script, err := d.script(filename)
	if err != nil {
		return nil, err
	}

	db, err := d.connection.Create(ctx)
	if err != nil {
		return nil, err
	}
	defer d.connection.Close(db)

	rows, err := db.NamedQuery(*script, model)
	if err != nil {
		return nil, err
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rows, nil
}

func (d *Database) script(filename string) (*string, error) {
	script, isOk := d.Scripts[filename]
	if !isOk {
		bytes, err := ioutil.ReadFile(filepath.Join(ScriptPath, filename))
		if err != nil {
			return nil, err
		}

		script = string(bytes)
		d.Scripts[filename] = script
	}
	return &script, nil
}

func bindUser(rows *sqlx.Rows) (*models.User, error) {
	if !rows.Next() {
		return nil, errors.New("value not found")
	}

	var value models.User
	if err := rows.StructScan(&value); err != nil {
		return nil, err
	}
	return &value, nil
}
