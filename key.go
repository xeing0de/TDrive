package main

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "log"
    "os"

    "golang.org/x/crypto/argon2"
    "golang.org/x/term"
)

func generateSalt(size int) ([]byte, error) {
    salt := make([]byte, size)

    _, err := rand.Read(salt)
    if err != nil {
        return nil, err
    }

    return salt, nil
}

func deriveKeyFromPassword(password []byte, salt []byte) []byte {
    var time uint32 = 1
    var memory uint32 = 64 * 1024
    var threads uint8 = 4
    var keyLen uint32 = 32

    key := argon2.IDKey(
        password,
        salt,
        time,
        memory,
        threads,
        keyLen,
    )

    return key
}

func main() {
    fmt.Print("Enter password: ")

    password, err := term.ReadPassword(int(os.Stdin.Fd()))
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println()

    salt, err := generateSalt(16)
    if err != nil {
        log.Fatal(err)
    }

    key := deriveKeyFromPassword(password, salt)

    fmt.Println("Salt:", hex.EncodeToString(salt))
    fmt.Println("Key: ", hex.EncodeToString(key))
}
