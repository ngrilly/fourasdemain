serve:
	source ./env && dev_appserver.py --env_var MAILGUN_API_KEY=$$MAILGUN_API_KEY .

deploy:
	cd hugo && hugo
	source ./env && appcfg.py update -E MAILGUN_API_KEY:$$MAILGUN_API_KEY .
