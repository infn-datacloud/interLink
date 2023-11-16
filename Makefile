all: interlink vk sidecars

interlink:
	go build -o bin/interlink cmd/interlink/main.go

vk:
	go build -o bin/vk

sidecars:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/docker-sd cmd/sidecars/docker/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/slurm-sd cmd/sidecars/slurm/main.go

clean:
	rm -rf ./bin

all_ppc64le:
	GOOS=linux GOARCH=ppc64le go build -o bin/interlink cmd/interlink/main.go
	GOOS=linux GOARCH=ppc64le go build -o bin/vk
	GOOS=linux GOARCH=ppc64le go build -o bin/docker-sd cmd/sidecars/docker/main.go
	GOOS=linux GOARCH=ppc64le go build -o bin/slurm-sd cmd/sidecars/slurm/main.go

start_interlink_slurm:
	./bin/interlink &> ./logs/interlink.log &
	./bin/slurm-sd  &> ./logs/slurm-sd.log &

start_interlink_docker:
	./bin/interlink &> ./logs/interlink.log &
	./bin/docker-sd  &> ./logs/docker-sd.log &
