//go:generate sh -c "cat generic.go | genny gen \"_Config=Config _Something=Service\" > config_service.go"
//go:generate sh -c "cat generic.go | genny gen \"_Config=Config _Something=Context\" > config_context.go"
//go:generate sh -c "cat generic.go | genny gen \"_Config=Config _Something=User\" > config_user.go"
//go:generate sh -c "cat generic.go | genny gen \"_Config=Service _Something=Environment\" > config_service_environment.go"
package grpctl
