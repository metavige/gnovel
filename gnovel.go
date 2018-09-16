// 下載 ck101 小說用
package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"flag"
	"strings"

	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/op/go-logging"
)

type BookInfo struct {
	Subject string
	Title   string
	Author  string
}

type NovelUrlQueryData struct {
	BookID int
	Page   int
}

type BookPageData struct {
	BookID    int
	BookStart int
	BookEnd   int
	//BookSeq   string
}

type BookDoc struct {
	Page int
	Doc  *goquery.Document
}

var (
	log       = logging.MustGetLogger("gnovel")
	urlFormat = "https://ck101.com/forum.php?mod=viewthread&tid=%v&page=%d"
)

func main() {

	novelURL := flag.String("url", "", "url")
	flag.Parse()

	novelURLStr := *novelURL
	if novelURLStr == "" {
		fmt.Println("no url")
		os.Exit(0)
	}

	// init logger
	var format = logging.MustStringFormatter("%{level} %{message}")
	logging.SetFormatter(format)
	logging.SetLevel(logging.INFO, "gnovel")

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

	var err error

	doc := getWebDocument(novelURLStr)

	// 用正規表達式找出書名以及作者
	bookInfo := getBookInfo(doc)
	fmt.Printf("書名： %v\n", bookInfo.Subject)

	urlPageData := getNovelUrlInfo(novelURLStr)

	pageStart := urlPageData.Page
	pageData := BookPageData{urlPageData.BookID, pageStart, pageStart}
	pageData.BookEnd = getBookPageEnd(doc)

	fmt.Printf("準備處理資料頁數: %d - %d\n", pageData.BookStart, pageData.BookEnd)

	// Open File
	novelFile := fmt.Sprintf("%v - %v.txt", bookInfo.Author, bookInfo.Title)
	var f *os.File
	if _, err = os.Stat(novelFile); os.IsExist(err) {
		os.Remove(novelFile)
	}
	f, err = os.Create(novelFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	totalPage := pageData.BookEnd - pageData.BookStart
	var pageDocuments = make([]*goquery.Document, totalPage+2)

	start := time.Now()
	//ch := make(chan BookDoc)
	for page := 1; page <= totalPage; page++ {
		pageURL := fmt.Sprintf(urlFormat, pageData.BookID, page)

		//printPage(pageURL, f)
		//go download(page, pageURL, ch)
		bookDoc := download(page, pageURL)
		pageDocuments[page] = bookDoc.Doc
	}

	// print page
	for page := 1; page <= totalPage; page++ {
		fmt.Printf("處理第 %d 頁 ..... \n", page)
		printPage(pageDocuments[page], f)
	}

	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())

}

// 因為 goquery 有改介面，所以把 url -> document 的部分獨立成一個方法
func getWebDocument(url string) (doc *goquery.Document) {
	res, e := http.Get(url)
	if e != nil {
		log.Fatal(e)
	}
	defer res.Body.Close()

	doc, _ = goquery.NewDocumentFromReader(res.Body)
	return doc
}

// 抓出書名與作者
func getBookInfo(doc *goquery.Document) (info BookInfo) {

	doc.Find("h1#thread_subject").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		//band := s.Find("a").Text()
		title := s.Text()
		title = strings.Replace(title, "【", "[", 1)
		title = strings.Replace(title, "】", "]", 1)
		title = strings.Replace(title, "] ", "]", 1)
		title = strings.Replace(title, " (", "(", 1)
		title = strings.Replace(title, "（", "(", 1)
		title = strings.Replace(title, "）", ")", 1)
		title = strings.Replace(title, "] [", "][", 1)
		title = strings.Replace(title, "  ", " ", 1)
		fmt.Printf("Review %d: %s\n", i, title)

		fmt.Println("title -> " + title)
		//fmt.Println(title)
		reg1 := "(\\S+) *(作者[：:] *(\\S+))"
		//reg1 := "(\\[(\\S+)\\])* *(\\S+) *(作者：) *(\\S+) *(\\S+)"
		// reg2 := "(\\[(\\S+)])?\\[(\\S+)](\\S+) 作者：(\\S+)\\((\\S+)\\)"
		r := regexp.MustCompile(reg1)
		bookInfo := r.FindStringSubmatch(title)
		fmt.Println("bookInfo: ", bookInfo[0])
		// 用正規表達式找出書名以及作者
		info = BookInfo{bookInfo[1], bookInfo[1], bookInfo[3]}
	})

	// title = ""
	return info
}

// 抓出最後一頁
func getBookPageEnd(doc *goquery.Document) (pageEnd int) {
	pageEnd = 1
	doc.Find("div#postlist div.pgt div.pg a.last").Each(func(i int, s *goquery.Selection) {
		// title := s.Text()
		href, _ := s.Attr("href")

		urlPageData := getNovelUrlInfo(href)

		fmt.Printf("BookEnd: %v\n", urlPageData.Page)
		pageEnd = urlPageData.Page
	})

	return
}

// 將檔案的文章部分輸出成 TXT
func printPage(doc *goquery.Document, file *os.File) {
	re := regexp.MustCompile("post_([0-9]*)")
	doc.Find("div#postlist div.plhin").Each(func(i int, s *goquery.Selection) {
		id, exist := s.Attr("id")
		if exist {
			//fmt.Printf("id: %v \n", id)
			matchData := re.FindAllStringSubmatch(id, -1)
			if len(matchData) > 0 {
				postID := matchData[0][1]
				// fmt.Println(postID)
				query := fmt.Sprintf("td#postmessage_%s", postID)
				// fmt.Println("query: ", query)
				s.Find(query).Each(func(i int, s *goquery.Selection) {
					file.WriteString(s.Text())
				})
			}
		}
	})
}

// 將檔案轉換成 goquery.Document
func download(page int, url string) BookDoc {
	fmt.Println("Downloading " + url + " ...")

	res, e := http.Get(url)
	if e != nil {
		log.Fatal(e)
	}
	defer res.Body.Close()

	doc, _ := goquery.NewDocumentFromReader(res.Body)
	//ch <- BookDoc{page, doc}
	return BookDoc{page, doc}
}

func getNovelUrlInfo(href string) (data NovelUrlQueryData) {

	u, _ := url.Parse(href)
	m, _ := url.ParseQuery(u.RawQuery)

	data.BookID, _ = strconv.Atoi(m["tid"][0])
	data.Page, _ = strconv.Atoi(m["page"][0])

	return data
}
