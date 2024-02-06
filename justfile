set dotenv-load

default:
	@just --choose

clear:
	clear

up: clear fmt
	pulumi up -e

up-debug: clear fmt
	pulumi up -e -d

fmt: 
	go fmt
