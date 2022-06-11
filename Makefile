.PHONY: clean minio-setup

clean:
	rm -rf minio
	rm -rf data

minio-setup:
	wget https://dl.min.io/server/minio/release/linux-amd64/minio
	chmod +x minio
	mkdir -p ./data
	nohup ./minio server ./data --console-address ":9001" &