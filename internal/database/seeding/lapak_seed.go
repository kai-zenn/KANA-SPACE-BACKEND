package seeding

import (
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"KANA-SPACE-BACKEND/internal/modules/lapak" 
)

type categoryNode struct {
	Name     string
	Slug     string
	Branch   string
	Children []categoryNode
}

func SeedCategories(db *gorm.DB) error {
	categoryTemplates := []categoryNode{
		{
			Name:   "Bahan Baku Daur Ulang",
			Slug:   "bahan-baku-daur-ulang",
			Branch: lapak.CategoryBranchRawMaterial,
			Children: []categoryNode{
				{Name: "Kain Perca", Slug: "kain-perca", Branch: lapak.CategoryBranchRawMaterial},
				{Name: "Kardus", Slug: "kardus", Branch: lapak.CategoryBranchRawMaterial},
				{Name: "Kertas", Slug: "kertas", Branch: lapak.CategoryBranchRawMaterial},
				{Name: "Botol Plastik", Slug: "botol-plastik", Branch: lapak.CategoryBranchRawMaterial},
			},
		},
		{
			Name:   "Produk Hasil Daur Ulang",
			Slug:   "produk-hasil-daur-ulang",
			Branch: lapak.CategoryBranchFinishedGoods,
			Children: []categoryNode{
				{
					Name:   "Perabot",
					Slug:   "perabot",
					Branch: lapak.CategoryBranchFinishedGoods,
					Children: []categoryNode{
						{Name: "Meja & Bangku Cafe", Slug: "meja-bangku-cafe", Branch: lapak.CategoryBranchFinishedGoods},
						{Name: "Rak & Storage", Slug: "rak-storage", Branch: lapak.CategoryBranchFinishedGoods},
					},
				},
				{
					Name:   "Hiasan & Dekorasi",
					Slug:   "hiasan-dekorasi",
					Branch: lapak.CategoryBranchFinishedGoods,
					Children: []categoryNode{
						{Name: "Hiasan Lampu", Slug: "hiasan-lampu", Branch: lapak.CategoryBranchFinishedGoods},
						{Name: "Tatakan Gelas", Slug: "tatakan-gelas", Branch: lapak.CategoryBranchFinishedGoods},
					},
				},
				{Name: "Aksesori & Pernak-pernik", Slug: "aksesori-pernak-pernik", Branch: lapak.CategoryBranchFinishedGoods},
				{Name: "Perhiasan", Slug: "perhiasan", Branch: lapak.CategoryBranchFinishedGoods},
				{Name: "Paving Block", Slug: "paving-block", Branch: lapak.CategoryBranchFinishedGoods},
			},
		},
	}

	var seedNode func(node categoryNode, parentID *uuid.UUID) error
	seedNode = func(node categoryNode, parentID *uuid.UUID) error {
		var existing lapak.Category
		err := db.Where("slug = ?", node.Slug).First(&existing).Error

		var currentCatID uuid.UUID

		if errors.Is(err, gorm.ErrRecordNotFound) {
			newCat := lapak.Category{
				ID:       uuid.New(),
				Name:     node.Name,
				Slug:     node.Slug,
				Branch:   node.Branch, 
				ParentID: parentID, 
			}
			if err := db.Create(&newCat).Error; err != nil {
				return fmt.Errorf("gagal membuat kategori %s: %w", node.Name, err)
			}
			currentCatID = newCat.ID
			log.Printf("Berhasil menambahkan kategori: %s", node.Name)
		} else if err != nil {
			return err
		} else {
			currentCatID = existing.ID
			log.Printf("Kategori %s sudah terdaftar, skip...", node.Name)
		}

		for _, child := range node.Children {
			if err := seedNode(child, &currentCatID); err != nil {
				return err
			}
		}

		return nil
	}

	for _, rootNode := range categoryTemplates {
		if err := seedNode(rootNode, nil); err != nil {
			return err
		}
	}

	return nil
}
