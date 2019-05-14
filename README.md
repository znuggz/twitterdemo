# twitterdemo
The program spins up 10 threads at a time to get tweets that contain the hashtag #IoT.
each threads tries to get 100 tweets, though the actual number returned by the free access Tweeter API does not garantee that requested number will actually get delivered.

as each thread gets unique tweets it writes them to a CSV file. Once 2000 unique tweets are collected, the channel is closed and the program stops.

some caveats, tweeters' standard (free) development account only allows 180 http requests in a 15 minute window, if the number of requests is reached, the program sleeps for 15 minutes and keeps trying until 2000 unique tweets are collected.

additionally, each tweet structure is composed of a series of top level items such as ID, text, language, etc, as well some substructures that hold user information, coordinate information, etc. The program separates the top level items and subtructs with commas, but does not parse out the substructure, it simply writes them to file as another column. 

before building, three open source go packages need to installed:
go get github.com/coreos/pkg/flagutil
go get github.com/dghubble/go-twitter/twitter
go get github.com/dghubble/oauth1

to install: go install

to run: cd $GOHOME/bin/ && ./tweeterdemo

to view file: less /tmp/tweets.csv
