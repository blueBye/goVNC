watch:
	watch -n 1 ls -lh ./FBS
build:
	sudo docker compose up -d --build
logs:
	sudo docker compose logs -f vnc_recorder
finit-logs:
	sudo docker compose logs vnc_recorder
exec:
	sudo docker compose exec -it vnc_recorder sh
down:
	sudo docker compose down
pull:
	sudo docker pull golang:1.22-alpine3.19
clean:
	sudo rm -f FBS/*
prune:
	sud docker builder prune
debug: clean down
	sleep 1
	sudo docker compose up -d --build
	sleep 1
	curl http://127.0.0.1:8090/record -X POST --data '{"server":"ed26df05-df05-4048-b02d-dd67ee401138","duration":20,"count":10}' -H "Accept: application/json"
log-file:
	sudo docker compose logs vnc_recorder > goVNC/1.log

