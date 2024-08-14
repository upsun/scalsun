package api

import (
	"github.com/upsun/lib-sun/entity"
	"github.com/upsun/scalsun/internal/logic"
)

func ScalingInstance(projectContext entity.ProjectGlobal) {
	logic.ScalingInstance(projectContext)
}
