default: frontend service

service:
	$(MAKE) -wC service default

proto: protocol

protocol:
	$(MAKE) -wC proto buf

frontend:
	$(MAKE) -wC frontend

test:
	$(MAKE) -wC service test
	$(MAKE) -wC frontend test

lint:
	$(MAKE) -wC service lint

run:
	$(MAKE) -wC service run

dev-frontend:
	$(MAKE) -wC frontend run

.PHONY: service protocol frontend test lint run dev-frontend
