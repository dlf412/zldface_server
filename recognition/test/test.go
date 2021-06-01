package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"time"
	reg "zldface_server/recognition"
)

func main() {
	eng, err := reg.NewEngine()
	if err != nil {
		log.Fatal(err)
	}
	features := map[string]interface{}{}

	if f, err := os.Open("test_features"); err == nil {
		dec := gob.NewDecoder(f)
		dec.Decode(&features)
		defer f.Close()
	} else {
		fmt.Println("开始提取特征...")
		bT := time.Now() // 开始时间
		for i := 1; i <= 50; i++ {
			pic := fmt.Sprintf("face/%d.jpg", i)

			if face, err := eng.DetectFace(pic); err == nil {
				feature, _ := eng.ExtractFeatureByteArr(face)
				features[pic] = feature
			}
		}
		eT := time.Since(bT) // 从开始到当前所消耗的时间
		fmt.Println("Extract 50 face cost time: ", eT)

		wf, _ := os.Create("test_features")
		defer wf.Close()
		encoder := gob.NewEncoder(wf)
		encoder.Encode(features)
	}

	println(len(features))
	bT := time.Now()
	// 开始时间
	for _, v := range features {
		eng.SearchN(v, features, 3, 0.75, 0.9)
	}
	eT := time.Since(bT) // 从开始到当前所消耗的时间
	fmt.Println("search 50 face cost time: ", eT)
}
