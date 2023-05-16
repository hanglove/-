package havenunit

import (
	"crypto/tls"
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/gomail.v2"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

var TimeLocal, _ = time.LoadLocation("Asia/Shanghai")

//保留2位小数
func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

//格式化千分位
func FmateFloatToStr(target float64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%.2f", target)

}

//获取两个日期中的所有日期

// GetFirstDateOfWeek 获取本周周日的日期UTC
func GetFirstSundayDateOfWeekUTC() (weekMonday string, weekStartDate time.Time) {
	now := time.Now().UTC()

	offset := int(time.Sunday - now.Weekday())
	if offset > 0 {
		offset = -6
	}

	weekStartDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	weekMonday = weekStartDate.Format("2006-01-02")
	return
}

//GetFirstDateOfWeek 获取本周周一的日期
func GetFirstDateOfWeek() (weekMonday string) {
	now := time.Now()

	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}

	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	weekMonday = weekStartDate.Format("2006-01-02")
	return
}

//GetLastWeekFirstDate 获取上周的周一日期
func GetLastWeekFirstDate() string {
	thisWeekMonday := GetFirstDateOfWeek()
	TimeMonday, _ := time.Parse("2006-01-02", thisWeekMonday)
	lastWeekMonday := TimeMonday.AddDate(0, 0, -7)
	weekMonday := lastWeekMonday.Format("2006-01-02")
	return weekMonday
}

//判断时间是当年的第几周
func WeekByDate(t time.Time) int {
	yearDay := t.YearDay()
	yearFirstDay := t.AddDate(0, 0, -yearDay+1)
	firstDayInWeek := int(yearFirstDay.Weekday())

	//今年第一周有几天
	firstWeekDays := 1
	if firstDayInWeek != 0 {
		firstWeekDays = 7 - firstDayInWeek + 1
	}
	var week int
	if yearDay <= firstWeekDays {
		week = 1
	} else {
		week = (yearDay-firstWeekDays)/7 + 2
	}
	//return fmt.Sprintf("%d第%d周", t.Year(), week)
	return week
}

//parameter timerange= 2020-12-12～2020-12-13
//return dateStrs []string
// h获取两个日期间的日期
func GetDateStr(timerange string) (dateStrs []string) {
	dateStrs = []string{}
	layout := "2006-01-02"
	timerange = strings.ReplaceAll(timerange, " ", "")
	//fmt.Println("timerange", timerange)
	dates := strings.Split(timerange, "~")
	if len(timerange) > 1 {
		start, err := time.ParseInLocation(layout, dates[0], TimeLocal)
		if err != nil {
			log.Println("ParseInLocation err", err)
			return
		}
		//fmt.Println("start", start)
		end, err := time.ParseInLocation(layout, dates[1], TimeLocal)
		if err != nil {
			log.Println("ParseInLocation err", err)
			return
		}
		//fmt.Println("end", end)
		day := end.Sub(start).Hours() / 24
		for i := 0; i < int(day+1); i++ {
			temp := start.AddDate(0, 0, i).Format(layout)
			dateStrs = append(dateStrs, temp)
		}
	}
	return
}

//parameter timerange= 2020-12-12 15～2020-12-13 15
//return dateStrs []string
// h获取两个日期间的日期 小时
func GetDateStrHour(timerange string) (dateStrs []string) {
	dateStrs = []string{}
	layout := "2006-01-02 15"
	//timerange = strings.ReplaceAll(timerange, " ", "")
	//fmt.Println("timerange", timerange)
	dates := strings.Split(timerange, "~")
	if len(timerange) > 1 {
		start, err := time.ParseInLocation(layout, dates[0]+" 00", TimeLocal)
		if err != nil {
			log.Println("ParseInLocation err", err)
			return
		}
		//fmt.Println("start", start)
		end, err := time.ParseInLocation(layout, dates[1]+" 23", TimeLocal)
		if err != nil {
			log.Println("ParseInLocation err", err)
			return
		}
		//fmt.Println("end", end)
		day := end.Sub(start).Hours()
		for i := 0; i < int(day+1); i++ {
			temp := start.Add(time.Duration(i) * time.Hour).Format(layout)
			dateStrs = append(dateStrs, temp)
		}
	}
	return
}

// Pool goroutine Pool
type Pool struct {
	queue chan int
	wg    *sync.WaitGroup
}

// New 新建一个协程池
func New(size int) *Pool {
	if size <= 0 {
		size = 1
	}
	return &Pool{
		queue: make(chan int, size),
		wg:    &sync.WaitGroup{},
	}
}

// Add 新增一个执行
func (p *Pool) Add(delta int) {
	// delta为正数就添加
	for i := 0; i < delta; i++ {
		p.queue <- 1
	}
	// delta为负数就减少
	for i := 0; i > delta; i-- {
		<-p.queue
	}
	p.wg.Add(delta)
}

// Done 执行完成减一
func (p *Pool) Done() {
	<-p.queue
	p.wg.Done()
}

func (p *Pool) Wait() {
	p.wg.Wait()
}

//发送邮件
func SendMail(host, fromUser, pass string, toUser []string, subject string, cc []string, filenameList []string, body string) error {
	// NewEmail返回一个email结构体的指针
	m := gomail.NewMessage()
	// 发件人
	m.SetHeader("From", "notice@iadbrain.com")
	// 收件人(可以有多个)
	m.SetHeader("To", toUser...)
	// 抄送人(可以有多个)
	m.SetHeader("Cc", cc...)
	// 邮件主题
	m.SetHeader("Subject", subject)
	//e.Subject = subject
	// html形式的消息
	m.SetBody("text/html", body)
	// 以路径将文件作为附件添加到邮件中
	for _, filename := range filenameList {
		var name string
		if strings.Contains(filename, "xlsx") {
			name = strings.ReplaceAll(filename, "./OutputFile/", "")
		} else if strings.Contains(filename, "zip") {
			name = "DataReport.zip"
		}
		//m.Attach("./OutputFile/【AsiaInnovations-Uplive】-报告.xlsx")
		//name := strings.ReplaceAll(filename,"./OutputFile/","")
		//name = strings.ReplaceAll(filename,"OutputFile/","")
		//h := map[string][]string{"Content-Transfer-Encoding": {"quoted-printable"}}
		m.Attach(filename, gomail.Rename(name))
	}
	// 发送邮件(如果使用QQ邮箱发送邮件的话，passwd不是邮箱密码而是授权码) 465
	d := gomail.NewDialer(host, 994, fromUser, pass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	d.SSL = true
	err := d.DialAndSend(m)
	//buf := new(bytes.Buffer)
	//_, _ = m.WriteTo(buf)
	//fmt.Println(buf.String())
	//Logs.Info(buf.String())
	//log.Panic("断点")

	return err
}
