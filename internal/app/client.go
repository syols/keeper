package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"keeper/config"
	"keeper/internal/models"
	"keeper/internal/pkg"
	pb "keeper/internal/rpc/proto"
)

// Client struct
type Client struct {
	ctx        context.Context
	authorizer pkg.Authorizer
	client     pb.KeeperClient
	database   pkg.Database
	settings   config.Config
}

func NewClient(ctx context.Context, settings config.Config) (Client, error) {
	uri := fmt.Sprintf(":%d", settings.Client.Address.Port)
	conn, err := grpc.Dial(uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return Client{}, err
	}

	dbConn := pkg.NewDatabaseURLConnection(*settings.Client.DatabaseConnectionString)
	database, err := pkg.NewDatabase(dbConn)
	if err := database.Truncate(ctx); err != nil {
		return Client{}, err
	}

	return Client{
		ctx:      ctx,
		client:   pb.NewKeeperClient(conn),
		database: database,
		settings: settings,
	}, nil
}

func (s *Client) Run() error {
	value := s.choose("1 - Sign in\n2 - Sign up")
	user := models.User{
		Username: s.getValue("User:"),
		Password: s.getValue("Password:"),
	}
	token, err := s.requestToken(value, user)
	if err != nil {
		return err
	}
	if err := s.database.Register(s.ctx, &user); err != nil {
		return err
	}

	if err := s.syncDatabase(err, user.Username, token); err != nil {
		return err
	}

	value = s.choose("1 - Show data\n2 - Write data")
	metadata := s.getValue("Metadata:")
	switch value {
	case "1":
		err := s.showData(metadata, user)
		if err != nil {
			return err
		}
		break
	case "2":
		if err := s.writeData(token, metadata); err != nil {
			return err
		}
		break
	}
	return nil
}

func (s *Client) showData(metadata string, user models.User) error {
	fmt.Println(metadata)
	records, err := s.database.UserRecords(s.ctx, user.Username)
	if err != nil {
		return err
	}
	for _, r := range records {
		fmt.Println(r)
	}
	return nil
}

func (s *Client) writeData(token *pb.SignInResponse, metadata string) error {
	v := pb.Record{
		AccessToken: token.Access,
		Metadata:    metadata,
		DetailType:  s.getValue("Metadata:"),
	}
	if v.DetailType == models.TextType {
		models.TextDetails{Data: s.getValue("Text:")}.SetPrivateData(&v)
	}
	if v.DetailType == models.LoginType {
		models.LoginDetails{Login: s.getValue("Login:"),
			Password: s.getValue("Password:")}.SetPrivateData(&v)
	}
	if v.DetailType == models.CardType {
		atoi, err := strconv.Atoi(s.getValue("Cvc:"))
		if err != nil {
			return err
		}
		models.CardDetails{
			Number:     s.getValue("Number:"),
			Cardholder: s.getValue("Cardholder:"),
			Cvc:        uint32(atoi),
		}.SetPrivateData(&v)
	}
	_, err := s.client.AddRecord(s.ctx, &v)
	if err != nil {
		return err
	}
	return nil
}

func (s *Client) choose(option string) string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Choose an option:")
	fmt.Println(option)
	fmt.Println("---------------------")
	scanner.Scan()
	return scanner.Text()
}

func (s *Client) syncDatabase(err error, username string, token *pb.SignInResponse) error {
	fmt.Println("Sync")
	syncClient, err := s.client.SyncRecord(s.ctx, token.Access)
	if err != nil {
		return err
	}

	errs := make(chan error, 1)
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				value, err := syncClient.Recv()
				if err == io.EOF {
					return
				}
				if err != nil {
					errs <- err
				}

				record := models.NewRecord(value)
				if err := s.database.SaveRecord(s.ctx, username, *record); err != nil {
					return
				}
			}
		}
	}()
	fmt.Println("OK")
	return nil
}

func (s *Client) requestToken(value string, user models.User) (*pb.SignInResponse, error) {
	request := user.SignInRequest()
	switch value {
	case "1":
		fmt.Println("Sign in")
		return s.client.Authorization(s.ctx, &request)
	case "2":
		fmt.Println("Sign up")
		return s.client.Registration(s.ctx, &request)
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
