package main

import "net/http"
import "net/http/cookiejar"
import "github.com/PuerkitoBio/goquery"
import "io/ioutil"
import "os"
import "os/exec"
import "net/url"
import "strings"
import "log"
import "strconv"

var client = http.Client{}

var uid = "15211010"
var pwd = "326526"

var week = "09"
var weekday = "4"
var time = 2 //1 for morning, 2 for afternoon and 3 for night
var experiment = "1021"

func isExperiment(row *goquery.Selection, experiment string) bool {
	if strings.Index(row.Children().First().Children().First().Text(), experiment) != -1 {
		return true
	} else {
		return false
	}
}

func login() {
	client.Get("http://phylab.buaa.edu.cn/logout.php")
	client.Jar, _ = cookiejar.New(nil)
	_, err := client.Get("http://phylab.buaa.edu.cn/index.php")
	if err != nil {
		log.Println(err)
		login()
		return
	}
	resp, err := client.Get("http://phylab.buaa.edu.cn/checkcode.php")
	if err != nil {
		log.Println(err)
		login()
		return
	}
	captchaData, _ := ioutil.ReadAll(resp.Body)
	ioutil.WriteFile("captcha.png", captchaData, os.ModePerm)
	captchaOutputData, _ := exec.Command("tesseract", "captcha.png", "stdout").Output()
	captchaOutputString := strings.Replace(string(captchaOutputData), " ", "", 0)
	client.PostForm("http://phylab.buaa.edu.cn/login.php", url.Values{
		"txtUid": {uid},
		"txtPwd": {pwd},
		"txtChk": {captchaOutputString},
	})
}

func main() {
	login()
	for {
		resp, err := client.Get("http://phylab.buaa.edu.cn/elect.php")
		if err != nil {
			log.Println(err)
			continue
		}
		d, _ := ioutil.ReadAll(resp.Body)
		if strings.HasPrefix(string(d), "<script") {
			log.Println("Log in failed")
			login()
			continue
		}
		//t.Sleep(30*t.Second + t.Duration(rand.Int31n(10))*t.Second)
		client.Get("http://phylab.buaa.edu.cn/elect.php?type=0&step=1&eid=bas" + week)
		client.PostForm("http://phylab.buaa.edu.cn/elect.php?type=0&step=2&eid=bas"+week, url.Values{
			"iknow": {"1"},
		})
		resp, err = client.PostForm("http://phylab.buaa.edu.cn/elect.php?type=0&step=2&eid=bas"+week, url.Values{
			"iknow":   {"1"},
			"weekday": {weekday},
		})
		if err != nil {
			log.Println(err)
			continue
		}
		doc, _ := goquery.NewDocumentFromResponse(resp)
		row := doc.Find("body > table:nth-child(6) > tbody > tr > td > form > table:nth-child(1) > tbody > tr > td > table > tbody > tr:nth-child(2)")
		for row.Length() > 0 && !isExperiment(row, experiment) {
			if row.Length() != 1 {
				log.Fatalln("Strange Behaviour having ", row.Length(), " nodes")
			}
			row = row.Next()
		}
		if row.Length() == 0 {
			log.Println("No such experiment or site down!")
			continue
		}
		col := row.Children().First()
		for i := 0; i < time; i++ {
			col = col.Next()
		}
		if len(col.Text()) < 8 {
			log.Println("Experiment not ready")
			continue
		} else if col.ChildrenFiltered("input").Length() == 0 {
			log.Fatalln("Experiment full!")
		} else {
			_, err = client.PostForm("http://phylab.buaa.edu.cn/elect.php?type=0&step=3&eid=bas"+week, url.Values{
				"Result": {experiment + week + weekday + strconv.FormatInt(int64(time), 10)},
			})

			for err != nil {
				_, err = client.PostForm("http://phylab.buaa.edu.cn/elect.php?type=0&step=3&eid=bas"+week, url.Values{
					"Result": {experiment + week + weekday + strconv.FormatInt(int64(time), 10)},
				})
			}

			log.Println("Done! Please check it out on the site yourself")

			//alert you with lots of notepads
			for i := 0; i < 10; i++ {
				exec.Command("notepad").Run()
			}

			return
		}

	}
}
