package app

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/syols/keeper/config"
	"github.com/syols/keeper/internal/models"
	"github.com/syols/keeper/internal/pkg"
	pb "github.com/syols/keeper/internal/rpc/proto"
)

type Choices map[int]string

// Client struct
type Client struct {
	client   pb.KeeperClient
	database pkg.Database
	settings config.Config
}

func NewClient(settings config.Config) (Client, error) {
	conn, err := grpcClient(settings)
	if err != nil {
		return Client{}, err
	}

	dbConn := pkg.NewDatabaseURLConnection(*settings.Client.DatabaseConnectionString)
	database, err := pkg.NewDatabase(dbConn)
	if err != nil {
		return Client{}, err
	}

	return Client{
		client:   pb.NewKeeperClient(conn),
		database: database,
		settings: settings,
	}, nil
}

func grpcClient(settings config.Config) (*grpc.ClientConn, error) {
	uri := fmt.Sprintf(":%d", settings.Client.Address.Port)
	if settings.Client.Certificate == nil {
		// Added to simplify debugging, environment variables are supported and TLS
		return grpc.Dial(uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	cert := []byte(*settings.Client.Certificate)
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(cert) {
		return nil, errors.New("incorrect certificates")
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            cp,
	}
	conn, err := grpc.Dial(uri, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *Client) Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	s.shutdown(ctx)

	err := s.database.Truncate(ctx)
	if err != nil {
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

	err = s.database.Register(ctx, &user)
	if err != nil {
		return err
	}

	err = s.syncDatabase(ctx, user.Username, token)
	if err != nil {
		return err
	}

	err = s.getUserDetails(ctx, user, token)
	if err != nil {
		return err
	}
	return nil
}

func (s *Client) getUserDetails(ctx context.Context, user models.User, token *pb.SignInResponse) error {
	value := s.choose(Choices{1: "Show data", 2: "Write data"})
	metadata := s.getValue("Metadata:")
	switch value {
	case "1":
		err := s.showData(ctx, metadata, user)
		if err != nil {
			return err
		}
	case "2":
		err := s.writeData(ctx, token, metadata)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Client) showData(ctx context.Context, metadata string, user models.User) error {
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
	record := pb.Record{
		AccessToken: token.Access,
		Metadata:    metadata,
		DetailType:  strings.ToUpper(s.getValue("DetailType:")),
	}
	switch record.DetailType {
	case models.TextType:
		models.TextDetails{Data: s.getValue("Text:")}.SetPrivateData(&record)
	case models.LoginType:
		models.LoginDetails{
			Login:    s.getValue("Login:"),
			Password: s.getValue("Password:")}.SetPrivateData(&record)
	case models.CardType:
		value, err := strconv.Atoi(s.getValue("Cvc:"))
		if err != nil {
			return err
		}

		models.CardDetails{
			Number:     s.getValue("Number:"),
			Cardholder: s.getValue("Cardholder:"),
			Cvc:        uint32(value),
		}.SetPrivateData(&record)
	default:
		return errors.New("incorrect detailType")
	}

	if _, err := s.client.AddRecord(ctx, &record); err != nil {
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
				if record != nil {
					err := s.database.SaveRecord(ctx, username, *record)
					if err != nil {
						return
					}
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
