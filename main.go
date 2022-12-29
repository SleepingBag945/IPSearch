package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func stringIpToInt(ipstring string) int {
	ipSegs := strings.Split(ipstring, ".")
	var ipInt int = 0
	var pos uint = 24
	for _, ipSeg := range ipSegs {
		tempInt, _ := strconv.Atoi(ipSeg)
		tempInt = tempInt << pos
		ipInt = ipInt | tempInt
		pos -= 8
	}
	return ipInt
}

func queryInfoByIP(ipStr string, db *sql.DB) {
	var inetnum string
	var netname string
	var country string
	var descr string
	var status string
	var last_modified string
	ip := stringIpToInt(ipStr)
	query := "select inetnum,netname,country,descr,status,\"last-modified\" from ipseg where start <= %d and end >= %d order by end-start ASC limit 0,1;"
	query = fmt.Sprintf(query, ip, ip)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&inetnum, &netname, &country, &descr, &status, &last_modified)
		if err != nil {
			log.Fatal(err)
		}
	}
	if brief {
		fmt.Printf("%s => %s\n", ipStr, inetnum)
	} else {
		fmt.Println("IP:", ipStr)
		fmt.Println("IP段:", inetnum)
		fmt.Println("名称:", netname)
		fmt.Println("描述:", descr)
		fmt.Println("国家:", country)
		fmt.Println("状态:", status)
		fmt.Println("最后修改:", last_modified)
		fmt.Println("")
	}

}

func queryInfoByKey(keys []string, db *sql.DB) {
	query := "select inetnum,netname,descr from ipseg where "
	descrQuery := "("
	for _, key := range keys {
		descrQuery += fmt.Sprintf("descr like \"%%%s%%\" and ", key)
	}
	descrQuery = descrQuery[:len(descrQuery)-5] + ") "
	query += descrQuery
	query += "or "
	netnameQuery := "("
	for _, key := range keys {
		netnameQuery += fmt.Sprintf("netname like \"%%%s%%\" and ", key)
	}
	netnameQuery = netnameQuery[:len(netnameQuery)-5] + ")"
	query += netnameQuery + ";"

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var inetnum string
		var netname string
		var descr string
		err = rows.Scan(&inetnum, &netname, &descr)
		if err != nil {
			log.Fatal(err)
		}
		count += 1
		fmt.Println("序号:", count)
		fmt.Println("IP段:", inetnum)
		fmt.Println("名称:", netname)
		fmt.Println("描述:", descr, "\n")
		// 限制数量
		if count > 2000 {
			break
		}
	}
}

var ipInput string
var ipFileName string
var key string
var brief bool

func main() {
	db, err := sql.Open("sqlite3", "IP.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	flag.StringVar(&ipInput, "i", "", "IP,用于查询所属IP段 -i 222.222.222.222 | 111.111.111.111,222.222.222.222")
	flag.StringVar(&ipFileName, "if", "", "IP文件，一行一个 -if ips.txt")
	flag.StringVar(&key, "k", "", "IP段关键词,多个关键词用,隔开 -k hangzhou,gov")
	flag.BoolVar(&brief, "b", false, "简要输出模式,在使用ip查询ip段时以 ip => ip段 的格式输出 -b")
	flag.Parse()

	//有指定单个ip的
	if ipInput != "" {
		ipList := []string{}
		if strings.Contains(ipInput, ",") {
			ipList = strings.Split(ipInput, ",")
			for _, each := range ipList {
				address := net.ParseIP(each)
				if address == nil {
					log.Fatal("-i 参数错误")
				}
			}
		} else {
			address := net.ParseIP(ipInput)
			if address == nil {
				log.Fatal("-i 参数错误")
			} else {
				ipList = append(ipList, ipInput)
			}
		}
		for _, each := range ipList {
			queryInfoByIP(each, db)
		}
		return
	}

	// 有指定ip文件的
	if ipFileName != "" {
		_, err = os.Stat(ipFileName)
		if err != nil {
			log.Fatal(err)
		}
		ips, err := os.ReadFile(ipFileName)
		if err != nil {
			log.Fatal(err)
		}
		tmpIpList := string(ips)
		ipList := []string{}
		for _, each := range strings.Split(tmpIpList, "\n") {
			if each == "" {
				continue
			}
			address := net.ParseIP(each)
			if address == nil {
				log.Fatal("-if 参数指向文件内容不合法")
			} else {
				ipList = append(ipList, each)
			}
		}
		for _, each := range ipList {
			queryInfoByIP(each, db)
		}
		return
	}

	// 有key存在
	if key != "" {
		keyList := strings.Split(key, ",")
		for _, each := range keyList {
			if each == "" {
				log.Fatal("关键词为空")
			}
		}
		queryInfoByKey(keyList, db)
		return
	}

}
