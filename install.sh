echo "building time ..."
go build -o ./bin/time time/main.go; sudo cp ./bin/time /usr/local/bin/time
echo "time updated"

echo "building timer ..."
go build -o ./bin/timer timer/main.go; sudo cp ./bin/timer /usr/local/bin/timer
echo "timer updated"





echo "all updated"