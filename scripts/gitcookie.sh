touch ~/.gitcookies
chmod 0600 ~/.gitcookies

git config --global http.cookiefile ~/.gitcookies

tr , \\t <<\__END__ >>~/.gitcookies
go.googlesource.com,FALSE,/,TRUE,2147483647,o,git-snappytheturtle1202.gmail.com=1/4zg9qf0zy1bmnMpEnqT7IL-rcL5_y8XP7bpfYE5DzN0
go-review.googlesource.com,FALSE,/,TRUE,2147483647,o,git-snappytheturtle1202.gmail.com=1/4zg9qf0zy1bmnMpEnqT7IL-rcL5_y8XP7bpfYE5DzN0
__END__
