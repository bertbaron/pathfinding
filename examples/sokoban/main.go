// sokoban solver, work in progress
package main

const (
	empty byte = 0
	wall byte = 1
	block byte = 2
	goal byte = 4
	player byte = 5
)

type context struct {
	field [][]byte
}

type mainstate struct {

}

func parse(level string) context {

}
var level = `
   ####
####  ##
#   $  #
#  *** #
#  . . ##
## * *  #
 ##***  #
  # $ ###
  # @ #
  #####`

func main() {

}
