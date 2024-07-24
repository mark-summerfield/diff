module github.com/mark-summerfield/diff2

go 1.22.5

replace github.com/mark-summerfield/set => /home/mark/app/golib/set

replace github.com/mark-summerfield/ureal => /home/mark/app/golib/ureal

replace github.com/mark-summerfield/utext => /home/mark/app/golib/utext

require (
	github.com/mark-summerfield/set v1.0.0
	github.com/mark-summerfield/utext v1.0.0
)

require github.com/mark-summerfield/ureal v1.0.0 // indirect
