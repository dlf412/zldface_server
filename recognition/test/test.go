package main

import (
	"fmt"
	reg "zldface_server/recognition"
)

func main() {

	score, err := reg.CompareImg("1.jpg", "10.jpg")
	fmt.Println(score, err)

	score, err = reg.CompareImg("1.jpg", "1.jpg")
	fmt.Println(score, err)

	feature1, err := reg.ImageToFeature("1.jpg")
	feature2, err := reg.ImageToFeature("10.jpg")

	score1, err := reg.CompareFeature(feature1, feature2)
	score2, err := reg.CompareImgFeature("1.jpg", feature2)

	fmt.Println(score1, score2)
}
