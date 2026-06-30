install-docker:
	@command -v docker >/dev/null 2>&1; \
		if [ $$? -ne 0 ]; then \
	    	echo >&2 "Docker is not installed. Installing Docker..."; \
	    	curl -fsSL https://get.docker.com -o get-docker.sh; \
	    	sh get-docker.sh; \
	    	rm get-docker.sh; \
	    else \
            echo "docker already installed ✔"; \
        fi

install-docker-compose:
	@command -v docker-compose >/dev/null 2>&1; \
    	if [ $$? -ne 0 ]; then \
    	    echo >&2 "docker-compose is not installed. Installing docker-compose..."; \
    	    sudo curl -L "https://github.com/docker/compose/releases/download/1.29.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose; \
    	    sudo chmod +x /usr/local/bin/docker-compose; \
    	else \
    	    echo "docker-compose already installed ✔"; \
    	fi

install-sql-migrate:
	@command -v sql-migrate >/dev/null 2>&1; \
		if [ $$? -ne 0 ]; then \
	    	echo >&2 "sql-migrate is not installed. Installing sql-migrate..."; \
	    	go install github.com/rubenv/sql-migrate/...@latest; \
		else \
        	echo "sql-migrate already installed ✔"; \
        fi

migrate-up:
	sql-migrate up -env=production -config=internal/infra/config/db/dbconfig.yml

done:
	@echo -e "\nALL TASKS DONE ✅";


all: install-docker install-docker-compose install-sql-migrate migrate-up done

.PHONY: install-docker install-docker-compose install-sql-migrate migrate-up
