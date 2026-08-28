default: deps prod

deps:
	NO_UPDATE_NOTIFIER=true npm install --no-fund --no-audit

prod:
	npx vite build --emptyOutDir

run:
	npx vite

test:
	node --test rbacAccess.test.js

.PHONY: deps prod run test
