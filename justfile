set dotenv-load

default:
	@just --choose

clear:
	clear

up: clear fmt
	pulumi up -d -e

fmt: 
	go fmt
