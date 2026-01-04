# Description


- The code inside this folder let's you see and test how to create modules, and even replace the "remote" repository, "locally"using go.mod file to replace or redirect the dependency management on golang.

`````go
module example.com/hello

go 1.25.5
//replace is used here to redirect the search for the dependency to the folder next to "home" folder.
replace example.com/greetings => ./../greetings

require example.com/greetings v0.0.0-00010101000000-000000000000
`````