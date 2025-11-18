package main

import (
	"bufio"
	"fmt"
	"github.com/Vestamart/wb-l2/blob/main/l2_9/unpack"
	"os"
)

func main() {
	fmt.Print("Введите строку: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Убираем \n в конце
	input = input[:len(input)-1]

	out, err := unpack.Unpack(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Распаковано:", out)
}
