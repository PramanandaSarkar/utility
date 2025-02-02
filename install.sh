echo "building time ..."
go build -o ./bin/time time/main.go; sudo cp ./bin/time /usr/bin/time
echo "time updated"

echo "building timer ..."
go build -o ./bin/timer timer/main.go; sudo cp ./bin/timer /usr/bin/timer
echo "timer updated"





echo "all updated"