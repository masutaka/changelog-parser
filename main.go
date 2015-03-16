package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Changelog struct {
	// CAT
	// STAT
	all All
}

type All struct {
	entry map[string]Entry
}

type Entry struct {
	curid int
	eh    string // entry header
	item  map[int]Item
}

type Item struct {
	a   string   // author
	co  string   // content
	ho  string   // header
	cat []string // category
}

func (cl *Changelog) debug_print() {
	for ymd := range cl.all.entry {
		ent := cl.all.entry[ymd]
		fmt.Printf("============================================================\n")
		fmt.Printf("ENTRY ID: %s\n", ymd)

		for i := ent.curid; i >= 1; i-- {
			fmt.Printf("------------------------------------------------------------\n")
			fmt.Printf("ITEM ID: %s-%d\n", ymd, i)
			fmt.Printf("ITEM HEADER:>>>>%s<<<<\n", ent.item[i].ho)
			fmt.Printf("ITEM CATEGORY:%s\n", ent.item[i].cat)
			fmt.Printf("ITEM AUTHOR:>>>>%s<<<<\n", ent.item[i].a)
			fmt.Printf("ITEM CONTENT:>>>>%s<<<<\n", ent.item[i].co)
		}
	}
}

func new_changelog() *Changelog {
	cl := new(Changelog)
	cl.all.entry = map[string]Entry{}
	return cl
}

func (cl *Changelog) store_changelog_file(file_name string) bool {
	file, err := os.Open(file_name)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	re := regexp.MustCompile(`^(\d{4}-\d\d-\d\d)`)

	var entlines []string

	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			if len(entlines) > 0 {
				cl.store_entry(entlines)
			}
			entlines = nil
		}
		entlines = append(entlines, line)
	}
	if len(entlines) > 0 {
		cl.store_entry(entlines)
	}

	if scanner.Err() != nil {
		fmt.Println("scanner error")
		return false
	}

	return true
}

func (cl *Changelog) store_entry(linesp []string) {
	// Processing entry header
	eh, linesp := linesp[0], linesp[1:]

	re := regexp.MustCompile(`^((\d{4})-(\d\d)-(\d\d))(?:.*?\s\s)(.+)?`)
	submatch := re.FindSubmatch([]byte(eh))
	ymd := string(submatch[1])
	// y := string(submatch[2])
	// m := string(submatch[3])
	// d := string(submatch[4])
	user := string(submatch[5])

	cl.all.entry[ymd] = Entry{eh: ymd}

	user = strings.NewReplacer("<", "&lt;", ">", "&gt;").Replace(user)

	var ilines []string
	var items [][]string

	re2 := regexp.MustCompile(`^( {8}| {0,7}\t|)\* `)
	re3 := regexp.MustCompile(`^\s*$`)

	for _, line := range linesp {
		if re2.MatchString(line) {
			line = re2.ReplaceAllString(line, "")
			if len(ilines) > 0 && !re3.MatchString(ilines[0]) {
				items = append(items, ilines)
			}
			ilines = nil
		}
		ilines = append(ilines, line)
	}
	if len(ilines) > 0 && !re3.MatchString(ilines[0]) {
		items = append(items, ilines)
	}

	for i := len(items) - 1; i >= 0; i-- {
		cl.all.entry[ymd] = cl.store_item(cl.all.entry[ymd], items[i], ymd, user)
	}

	// foreach (reverse @items) {
	// 	cl.store_item()
	// }
}

func (cl *Changelog) store_item(entry Entry, linesp []string, ymd, user string) Entry {
	ih, linesp := linesp[0], linesp[1:]

	// item header - case 1: "* AAA: \n"
	// item header - case 2: "* AAA:\n"
	// item header - case 3: "* AAA: BBB\n"
	// item header - case 4: "* AAA\n"
	re := regexp.MustCompile(`:(\s.*)$`)
	var rest = ""
	if re.MatchString(ih) {
		tmp := re.FindSubmatch([]byte(ih)) // for case 1,2,3
		rest = string(tmp[1])
		rest = regexp.MustCompile(`^ +`).ReplaceAllString(rest, "")
	}

	cont := rest
	for _, hoge := range linesp {
		cont += hoge
	}

	if regexp.MustCompile(`^p:`).MatchString(ih) { // Ignoring private items
		return Entry{}
	}

	// item ID : Y in XXXX-XX-XX-Y
	entry.curid++

	// Processing item header
	// If 1st line doesn't have ": ", it will become item header.
	// var cat []string
	// ih = regexp.MustCompile(`(:|\s*)$`).ReplaceAllString(ih, "") // Triming trailing spaces and ":"
	// if regexp.MustCompile(`\s*\[(.+)\]$`).MatchString(ih) {      // category
	// 	regexp.MustCompile(`\s*\]\s*\[\s*`)
	// 	cat = split(/\s*\]\s*\[\s*/, $1);
	// }

	if entry.curid == 1 {
		entry.item = map[int]Item{}
	}

	// Processing item content
	ih = regexp.MustCompile(`^( {8}| {0,7}\t)`).ReplaceAllString(ih, "")
	ih = regexp.MustCompile(`\s+$`).ReplaceAllString(ih, "\n")
	ih = regexp.MustCompile(`\r`).ReplaceAllString(ih, "")

	// Storing item information in hash
	entry.item[entry.curid] = Item{
		ho:  ih,
		co:  cont,
		a:   user,
		cat: nil,
	}

	return entry
}

func main() {
	cl := new_changelog()
	cl.store_changelog_file("./ChangeLog.small")
	cl.debug_print()
}
