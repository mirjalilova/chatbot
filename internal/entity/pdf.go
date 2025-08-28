package entity

type DivorceRequest struct {
	CourtName          string `json:"court_name" binding:"required"`
	ClaimantFullName   string `json:"claimant_full_name" binding:"required"`
	ClaimantAddress    string `json:"claimant_address" binding:"required"`
	ClaimantPhone      string `json:"claimant_phone" binding:"required"`
	ClaimantEmail      string `json:"claimant_email" binding:"required"`
	RespondentFullName string `json:"respondent_full_name" binding:"required"`
	RespondentAddress  string `json:"respondent_address" binding:"required"`
	RespondentPhone    string `json:"respondent_phone" binding:"required"`
	RespondentEmail    string `json:"respondent_email" binding:"required"`
	FhdyoOffice        string `json:"fhdyo_office" binding:"required"`
	MarriageDate       string `json:"marriage_date" binding:"required"`
	CertificateNumber  string `json:"certificate_number" binding:"required"`
	ChildFullName      string `json:"child_full_name" binding:"required"`
	ChildBirthDate     string `json:"child_birth_date" binding:"required"`
	ChildFhdyo         string `json:"child_fhdyo" binding:"required"`
	ChildCertificate   string `json:"child_certificate" binding:"required"`
	DivorceReason      string `json:"divorce_reason" binding:"required"`
	ApplicationDate    string `json:"application_date" binding:"required"`
}

type CreatePdfCategory struct {
	Name string `json:"name" binding:"required"`
}

type UpdatePdfCategory struct {
	Id   string `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type CretatePdfCategoryItem struct {
	Name          string `json:"name" binding:"required"`
	PdfCategoryId string `json:"pdf_category_id" binding:"required"`
}

type UpdatePdfCategoryItem struct {
	Id            string `json:"id" binding:"required"`
	Name          string `json:"name" binding:"required"`
	PdfCategoryId string `json:"pdf_category_id" binding:"required"`
}

type ListPdfCategoryItem struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ListPdfCategory struct {
	Id    string                `json:"id"`
	Name  string                `json:"name"`
	Items []ListPdfCategoryItem `json:"items"`
}

type ApplicationRequired struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required"`
}

type ListApplicationRequired struct {
	Id   string `json:"id"`
	Text string `json:"text"`
	Type string `json:"type"`
}

type ApplicationItems struct {
	ApplicationItems []ListApplicationRequired `json:"application_requireds" binding:"required"`
}
