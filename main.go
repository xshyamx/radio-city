package main

import (
	"fmt"
)

func main() {

	urls := []string{
		"https://www.radiocity.in/radiocity/show-podcasts-tamil/Crime-Diary/153",
		"https://www.radiocity.in/radiocity/show-podcasts-hindi/Kissa-Crime-Ka/82",
	}

	rss, err := feedFromUrls(urls)
	if err != nil {
		panic(err)
	}
	out, err := writeFeed(rss)
	if err != nil {
		panic(err)
	}
	fmt.Println(out.String())
}
