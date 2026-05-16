package main

import (
    "encoding/ascii85"
    "fmt"
    "os"
)

func main() {
    if len(os.Args) != 3 {
        fmt.Println("Использование:")
        fmt.Println("go run main.go input_file output.txt")
        os.Exit(1)
    }

    inputPath := os.Args[1]
    outputPath := os.Args[2]

    data, err := os.ReadFile(inputPath)
    if err != nil {
        fmt.Println("Ошибка чтения файла:", err)
        os.Exit(1)
    }

    encoded := make([]byte, ascii85.MaxEncodedLen(len(data)))

    encodedLen := ascii85.Encode(encoded, data)
    encoded = encoded[:encodedLen]

    err = os.WriteFile(outputPath, encoded, 0644)
    if err != nil {
        fmt.Println("Ошибка записи файла:", err)
        os.Exit(1)
    }

    fmt.Println("Готово")
    fmt.Println("Входной файл:", inputPath)
    fmt.Println("Выходной файл:", outputPath)
    fmt.Println("Исходный размер:", len(data), "байт")
    fmt.Println("Размер ASCII85:", len(encoded), "символов")
}
