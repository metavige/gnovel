package main

import (
	//"os"

	//"github.com/PuerkitoBio/goquery"
	logging "github.com/op/go-logging"
	//"github.com/urfave/cli"
	"fmt"
	"github.com/PuerkitoBio/goquery"
)

var (
	log = logging.MustGetLogger("gnovel")
)

func main() {
	// init logger
	var format = logging.MustStringFormatter("%{level} %{message}")
	logging.SetFormatter(format)
	logging.SetLevel(logging.INFO, "gnovel")

	doc, err := goquery.NewDocument("https://ck101.com/thread-3397486-1-3.html")
	if err != nil {
		log.Fatal(err)
	}

	/*
	  1. 找出標題
	  2. 找出頁籤，並找到總共幾頁
	   用 pattern 抓出 url 的位置，主要是要知道每頁變化的位置
	   然後找出 pgt 位置，找到最後一頁，就可以用 range 來排出全部的網頁位置
	   接下來，依序把每一頁的網頁都抓下來

	   然後解析每頁的網頁

	   1. 找出 div$postlist 內所有 div
	   找出 id 是 post_XXXX 的
	   然後抓出 XXXX
	   		抓出底下 postmessage_XXXX 的 Text 就是每樓的文字內容

	   接下來就可以整合起來，輸出成 text

	*/
	doc.Find("h1#thread_subject").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		//band := s.Find("a").Text()
		title := s.Text()
		fmt.Printf("Review %d: %s\n", i, title)
	})

	// Print result list
	doc.Find("div#postlist div.pgt a").Each(func(i int, s *goquery.Selection) {
		title := s.Text()
		href, exist := s.Attr("href")
		if exist {

			//hrefs = append(hrefs, href)
			fmt.Printf("[%v] %v - %v\n", i, title, href)
		}
		//fmt.Printf("%v", title)
	})
	//
	//
	//app := cli.NewApp()
	//
	//app.Flags = []cli.Flag{
	//	cli.StringFlag{
	//		Name:  "url",
	//		Usage: "download url",
	//	},
	//}
	//
	//app.Run(os.Args)
}
