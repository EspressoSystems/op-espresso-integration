package main

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-service/espresso/hotshot"
)

func main() {
	l1 := "http://localhost:8545"
	hotshotAddr := "0xd710a67624ad831683c86a48291c597ade30f787"
	hotshot, err := hotshot.NewHotShotProvider(l1, hotshotAddr)
	fmt.Println(hotshot, err)
	roots, err := hotshot.GetCommitmentsFromHeight(0, 5)
	fmt.Println(roots, err)
}
