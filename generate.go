//go:generate sh -c "cat generic.go | genny gen \"_Config=Config _Something=Service\" > config_service.go"
//go:generate sh -c "cat generic.go | genny gen \"_Config=Config _Something=User\" > config_user.go"
package grpctl
