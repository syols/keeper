package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/syols/keeper/config"
	"github.com/syols/keeper/internal/models"
	"github.com/syols/keeper/internal/pkg"
	pb "github.com/syols/keeper/internal/rpc/proto"
)

type Choices map[int]string

// Client struct
type Client struct {
	authorizer pkg.Authorizer
	client     pb.KeeperClient
	database   pkg.Database
	settings   config.Config
}

func NewClient(settings config.Config) (Client, error) {
	uri := fmt.Sprintf(":%d", settings.Client.Address.Port)
	conn, err := grpc.Dial(uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return Client{}, err
	}

	dbConn := pkg.NewDatabaseURLConnection(*settings.Client.DatabaseConnectionString)
	database, err := pkg.NewDatabase(dbConn)
	return Client{
		client:   pb.NewKeeperClient(conn),
		database: database,
		settings: settings,
	}, nil
}

func (s *Client) Run(ctx context.Context) error {
	if err := s.database.Truncate(ctx); err != nil {
		return err
	}

	value := s.choose(Choices{1: "Sign in", 2: "Sign up"})
	user := models.User{
		Username: s.getValue("User:"),
		Password: s.getValue("Password:"),
	}
	token, err := s.requestToken(ctx, value, user)
	if err != nil {
		return err
	}
	if err := s.database.Register(ctx, &user); err != nil {
		return err
	}

	if err := s.syncDatabase(ctx, user.Username, token); err != nil {
		return err
	}

	value = s.choose(Choices{1: "Show data", 2: "Write data"})
	metadata := s.getValue("Metadata:")
	switch value {
	case "1":
		if err := s.showData(ctx, metadata, user); err != nil {
			return err
		}
		break
	case "2":
		if err := s.writeData(ctx, token, metadata); err != nil {
			return err
		}
		break
	}
	return nil
}

func (s *Client) showData(ctx context.Context, metadata string, user models.User) error {
	fmt.Println(metadata)
	records, err := s.database.UserRecords(ctx, user.Username)
	if err != nil {
		return err
	}
	for _, r := range records {
		fmt.Println(r)
	}
	return nil
}

func (s *Client) writeData(ctx context.Context, token *pb.SignInResponse, metadata string) error {
	v := pb.Record{
		AccessToken: token.Access,
		Metadata:    metadata,
		DetailType:  strings.ToUpper(s.getValue("DetailType:")),
	}
	switch v.DetailType {
	case models.TextType:
		models.TextDetails{Data: s.getValue("Text:")}.SetPrivateData(&v)
	case models.LoginType:
		models.LoginDetails{
			Login:    s.getValue("Login:"),
			Password: s.getValue("Password:")}.SetPrivateData(&v)
	case models.CardType:
		value, err := strconv.Atoi(s.getValue("Cvc:"))
		if err != nil {
			return err
		}
		models.CardDetails{
			Number:     s.getValue("Number:"),
			Cardholder: s.getValue("Cardholder:"),
			Cvc:        uint32(value),
		}.SetPrivateData(&v)
	default:
		return errors.New("incorrect detailType")
	}

	if _, err := s.client.AddRecord(ctx, &v); err != nil {
		return err
	}
	return nil
}

func (s *Client) choose(choices Choices) string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Choose an option:")
	for k, v := range choices {
		fmt.Printf("%v - %v\n", k, v)
	}
	fmt.Println("---------------------")
	scanner.Scan()
	return scanner.Text()
}

func (s *Client) syncDatabase(ctx context.Context, username string, token *pb.SignInResponse) error {
	fmt.Println("Sync")
	syncClient, err := s.client.SyncRecord(ctx, token.Access)
	if err != nil {
		return err
	}

	errs := make(chan error, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				value, err := syncClient.Recv()
				if err == io.EOF {
					return
				}
				if err != nil {
					errs <- err
					return
				}

				record := models.NewRecord(value)
				if record == nil {
					log.Println("Empty record")
				}

				if err := s.database.SaveRecord(ctx, username, *record); err != nil {
					return
				}
			}
		}
	}()
	fmt.Println("Sync OK!")
	return nil
}

func (s *Client) requestToken(ctx context.Context, value string, user models.User) (*pb.SignInResponse, error) {
	request := user.SignInRequest()
	switch value {
	case "1":
		return s.client.Authorization(ctx, &request)
	case "2":
		return s.client.Registration(ctx, &request)
	}
	return nil, errors.New("incorrect input")
}

func (s *Client) getValue(text string) string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(text)
	scanner.Scan()
	login := scanner.Text()
	return login
}

func (s *Client) shutdown(ctx context.Context) {
	go func() {
		<-ctx.Done()
		os.Exit(0)
	}()
}
