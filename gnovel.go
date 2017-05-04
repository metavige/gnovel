// 下載 ck101 小說用
package main

import (
	//"os"

	//"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	logging "github.com/op/go-logging"
	//"github.com/urfave/cli"
	"fmt"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

type BookInfo struct {
	Subject string
	Title   string
	Author  string
}

type BookPageData struct {
	BookID    string
	BookStart int
	BookEnd   int
	BookSeq   string
}

var (
	log        = logging.MustGetLogger("gnovel")
	url        = "https://ck101.com/thread-3397486-1-3.html"
	pageRegExp = regexp.MustCompile("thread-(\\d+)-(\\d+)-(\\d+).html")
	urlFormat  = "https://ck101.com/thread-%v-%d-%v.html"
	tmpFileDir = os.TempDir()
)

//"/Volumes/RamDisk/tmp"
func main() {
	// init logger
	var format = logging.MustStringFormatter("%{level} %{message}")
	logging.SetFormatter(format)
	logging.SetLevel(logging.INFO, "gnovel")

	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(tmpFileDir); os.IsNotExist(err) {
		os.Mkdir(tmpFileDir, os.ModePerm)
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

	// 用正規表達式找出書名以及作者
	bookInfo := getBookInfo(doc)
	fmt.Printf("書名： %v\n", bookInfo.Subject)

	// 找出最後一頁之後，用正規表達式，找出最後一頁的數字，跑回圈抓完文件
	if pageRegExp.Match([]byte(url)) {
		startPageMatch := pageRegExp.FindStringSubmatch(path.Base(url))
		// fmt.Printf("Url Match: %d", len(startPageMatch))
		pageStart, _ := strconv.Atoi(startPageMatch[2])
		pageData := BookPageData{startPageMatch[1], pageStart, pageStart, startPageMatch[3]}
		pageData.BookEnd = getBookPageEnd(doc)

		fmt.Printf("準備處理資料頁數: %d - %d\n", pageData.BookStart, pageData.BookEnd)

		start := time.Now()
		ch := make(chan string)
		for page := pageData.BookStart; page <= pageData.BookEnd; page++ {
			pageURL := fmt.Sprintf(urlFormat, pageData.BookID, page, pageData.BookSeq)

			// pringPage(pageURL)
			go download(pageURL, ch)
		}

		for page := pageData.BookStart; page <= pageData.BookEnd; page++ {
			fmt.Println(<-ch) // receive from channel ch
		}
		fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())

		// Cleanup Temp
		os.RemoveAll(tmpFileDir)
	} else {
		// 沒抓到最後一頁的網址
		fmt.Print("沒抓到最後一頁網址，不處理")
	}

}

// 抓出書名與作者
func getBookInfo(doc *goquery.Document) (info BookInfo) {

	doc.Find("h1#thread_subject").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		//band := s.Find("a").Text()
		title := s.Text()
		// fmt.Printf("Review %d: %s\n", i, title)

		r := regexp.MustCompile("\\[(\\S+)\\] (\\S+) 作者：(\\S+) \\((\\S+)\\)")
		bookInfo := r.FindStringSubmatch(title)
		// 用正規表達式找出書名以及作者
		info = BookInfo{bookInfo[0], bookInfo[2], bookInfo[3]}
	})

	// title = ""
	return info
}

// 抓出最後一頁
func getBookPageEnd(doc *goquery.Document) (pageEnd int) {
	bookEnd := 1
	doc.Find("div#postlist div.pgt a.last").Each(func(i int, s *goquery.Selection) {
		// title := s.Text()
		href, _ := s.Attr("href")
		// 找出最後一頁之後，用正規表達式，找出最後一頁的數字
		hrefMatch := pageRegExp.FindStringSubmatch(path.Base(href))
		fmt.Printf("BookEnd: %v", hrefMatch[2])
		bookEnd, _ = strconv.Atoi(hrefMatch[2])
	})

	return bookEnd
}

// 將檔案的文章部分輸出成 TXT
func pringPage(pageURL string) {
	doc, _ := goquery.NewDocument(pageURL)
	re := regexp.MustCompile("post_([0-9]*)")
	doc.Find("div#postlist div.plhin").Each(func(i int, s *goquery.Selection) {
		id, exist := s.Attr("id")
		if exist {
			//fmt.Printf("id: %v \n", id)
			matchData := re.FindAllStringSubmatch(id, -1)
			if len(matchData) > 0 {
				postID := matchData[0][1]
				fmt.Println(postID)
				// TODO: 解析內容，轉成文字
				// query := fmt.Sprintf("td#postmessage_%s", postID)
				// fmt.Println("query: ", query)
				// s.Find(query).Each(func(i int, s *goquery.Selection) {
				// 	fmt.Println(s.Text())
				// })
			}
		}
	})
}

// 下載檔案
func download(url string, ch chan<- string) {
	// fmt.Printf("下載：%v\n", pageURL)

	fmt.Println("Downloading " + url + " ...")
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprint(err) // send to channel ch
		return
	}

	destfile := fmt.Sprintf("%v/%v", tmpFileDir, path.Base(url))
	f, _ := os.Create(destfile)
	nbytes, err := io.Copy(f, resp.Body)
	resp.Body.Close() // don't leak resources
	if err != nil {
		ch <- fmt.Sprintf("while reading %s: %v", url, err)
		return
	}
	secs := time.Since(start).Seconds()
	ch <- fmt.Sprintf("%.2fs  %7d  %s", secs, nbytes, destfile)
}
