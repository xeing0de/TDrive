package main

import (
    "bufio"
    "context"
    "errors"
    "fmt"
    "os"
    "os/signal"
    "strconv"
    "strings"
	"tdrive/tgclient"

    "github.com/gotd/td/session"
    "github.com/gotd/td/telegram"
    "github.com/gotd/td/telegram/auth"
    "github.com/gotd/td/tg"
    "github.com/joho/godotenv"
	"golang.org/x/term"
	"github.com/gotd/contrib/middleware/floodwait"
)

type consoleAuth struct {
    phone  string
    reader *bufio.Reader
}

func (a *consoleAuth) Phone(ctx context.Context) (string, error) {
    return a.phone, nil
}

func (a *consoleAuth) Password(ctx context.Context) (string, error) {
    fmt.Print("Введите пароль 2FA, если Telegram попросил: ")
    passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
    if err != nil {
        return "", err
    }

    return string(passwordBytes), nil
}

func (a *consoleAuth) SignUp(ctx context.Context) (auth.UserInfo, error) {
    return auth.UserInfo{}, errors.New("sign up is disabled: use an existing Telegram account")
}

func (a *consoleAuth) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
    fmt.Print("Введите код из Telegram: ")
    text, err := a.reader.ReadString('\n')
    if err != nil {
        return "", err
    }

    return strings.TrimSpace(text), nil
}

func (a *consoleAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
    fmt.Println("Telegram требует принять Terms of Service.")
    fmt.Print("Введите yes для принятия: ")

    text, err := a.reader.ReadString('\n')
    if err != nil {
        return err
    }

    if strings.TrimSpace(strings.ToLower(text)) != "yes" {
        return errors.New("terms of service not accepted")
    }

    return nil
}

func main() {
    fmt.Println("Старт программы")
    _ = godotenv.Load()
	
    apiIDRaw := os.Getenv("API_ID")
    apiHash := os.Getenv("API_HASH")
    phone := os.Getenv("PHONE")
    sessionFile := os.Getenv("SESSION_FILE")

    if sessionFile == "" {
        sessionFile = "session.json"
    }

    if apiIDRaw == "" || apiHash == "" || phone == "" {
        fmt.Fprintln(os.Stderr, "Заполни API_ID, API_HASH и PHONE в .env")
        os.Exit(1)
    }

    apiID, err := strconv.Atoi(apiIDRaw)
    if err != nil {
        fmt.Fprintln(os.Stderr, "API_ID должен быть числом:", err)
        os.Exit(1)
    }

    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()

	waiter := floodwait.NewWaiter()

    client := telegram.NewClient(apiID, apiHash, telegram.Options{
        SessionStorage: &session.FileStorage{
            Path: sessionFile,
        },
		Middlewares: []telegram.Middleware{
        	waiter,
    	},
    })

    flow := auth.NewFlow(&consoleAuth{
        phone:  phone,
        reader: bufio.NewReader(os.Stdin),
    }, auth.SendCodeOptions{})

    err = waiter.Run(ctx , func(ctx context.Context) error {
		return client.Run(ctx, func(ctx context.Context) error {
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
            return err
        }

        self, err := client.Self(ctx)
        if err != nil {
            return err
        }

        username := self.Username
        if username == "" {
            username = "без username"
        }

        fmt.Printf("Подключение успешно: %s %s, @%s, id=%d\n",
            self.FirstName,
            self.LastName,
            username,
            self.ID,
        )
		api := tg.NewClient(client)
		/*group, err := tgclient.CreateSupergroup(ctx, api, "Test");
		if err != nil {
			return err
		}*/
		groups, err := tgclient.FindDiskGroups(ctx, api)
		if err != nil{
			return err
		}
		
		fmt.Println(groups[0].Title)
		
		forumid, err := tgclient.CreateTopic(ctx, client, groups[0].Channel, "отшельник")
		if err != nil{
			return err
		}
		fmt.Println(forumid)

		if err := tgclient.DeleteTopic(ctx, client, groups[0].Channel, forumid); err != nil{
			return err
		}
		
		count, err := tgclient.GetTopicMessageCount(ctx, api, groups[0].Peer, 1); 
		if err != nil {
			return err
		}
		fmt.Println(count)

		s := strings.Repeat("a", 4096)
		if err := tgclient.SendMessageToTopic(ctx, api, groups[0].Peer, 1, s); err != nil{
			return err
		}

        return nil
    })})

    if err != nil {
        fmt.Fprintln(os.Stderr, "Ошибка:", err)
        os.Exit(1)
    }
}
