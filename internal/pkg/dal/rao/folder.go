package rao

type SaveFolderReq struct {
	TargetID    string `json:"target_id"`
	TeamID      string `json:"team_id" binding:"required,gt=0"`
	ParentID    string `json:"parent_id"`
	Name        string `json:"name" binding:"required,min=1"`
	Method      string `json:"method"`
	Sort        int32  `json:"sort"`
	TypeSort    int32  `json:"type_sort"`
	Version     int32  `json:"version"`
	Description string `json:"description"`
	Source      int32  `json:"source"`
	SourceID    string `json:"source_id"`
	//Request  *Request `json:"request"`
	//Script   *Script  `json:"script"`
}

type SaveFolderResp struct {
}

type GetFolderReq struct {
	TeamID   string `form:"team_id"`
	TargetID string `form:"target_id"`
}

type GetFolderResp struct {
	Folder *Folder `json:"folder"`
}

type Folder struct {
	TargetID    string   `json:"target_id"`
	TeamID      string   `json:"team_id"`
	ParentID    string   `json:"parent_id"`
	Name        string   `json:"name"`
	Method      string   `json:"method"`
	Sort        int32    `json:"sort"`
	TypeSort    int32    `json:"type_sort"`
	Version     int32    `json:"version"`
	Request     *Request `json:"request"`
	Script      *Script  `json:"script"`
	Description string   `json:"description"`
}
