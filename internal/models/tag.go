package models

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CreateTagRequest — тело POST /tags
type CreateTagRequest struct {
	Name string `json:"name" validate:"required,max=50"`
}

// AttachTagsRequest — тело POST /manuals/{id}/tags
type AttachTagsRequest struct {
	TagIDs []int `json:"tag_ids" validate:"required,min=1,dive,gt=0"`
}
