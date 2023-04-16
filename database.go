package main

type IDatabase interface {
	GetSampleListOfCompaniesForIndustry(industry *string) (*[]ProcessedCompany, error)
}

type Database struct {
}
