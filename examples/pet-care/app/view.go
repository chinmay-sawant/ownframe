package app

import "github.com/chinmay-sawant/ownframe/examples/pet-care/model"

type View struct {
	Screen, Status, Error, Theme, Search, StorageWarning string
	BackupJSON, RestoreMessage                           string
	ConfirmRecord                                        string
	Email, Password, ConfirmPet                          string
	PetName, PetSpecies, PetBreed, PetBirth, PetNotes    string
	EditPetID                                            string
	Pets                                                 []model.Pet
	AllPets                                              []model.Pet
	Records                                              []RecordRow
	Recent                                               []RecordRow
	Selected                                             model.Pet
	HasSelected, SignedIn, Dark, ConfirmRestore          bool
}

type RecordRow struct {
	model.Record
	PetName, Status string
}
