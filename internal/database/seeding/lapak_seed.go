package seeding

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"KANA-SPACE-BACKEND/internal/modules/lapak"
	"KANA-SPACE-BACKEND/internal/modules/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
					Name:   "Tas & Aksesori",
					Slug:   "tas-aksesori",
					Branch: lapak.CategoryBranchFinishedGoods,
					Children: []categoryNode{
						{Name: "Tote Bag", Slug: "tote-bag", Branch: lapak.CategoryBranchFinishedGoods},
						{Name: "Dompet & Pouch", Slug: "dompet-pouch", Branch: lapak.CategoryBranchFinishedGoods},
					},
				},
				{
					Name:   "Home & Living",
					Slug:   "home-living",
					Branch: lapak.CategoryBranchFinishedGoods,
					Children: []categoryNode{
						{Name: "Sarung Bantal", Slug: "sarung-bantal", Branch: lapak.CategoryBranchFinishedGoods},
						{Name: "Keset", Slug: "keset", Branch: lapak.CategoryBranchFinishedGoods},
						{Name: "Kotak Penyimpanan", Slug: "kotak-penyimpanan", Branch: lapak.CategoryBranchFinishedGoods},
						{Name: "Rak Mini Organizer", Slug: "rak-mini-organizer", Branch: lapak.CategoryBranchFinishedGoods},
					},
				},
				{
					Name:   "Alat Tulis & Kantor",
					Slug:   "alat-tulis-kantor",
					Branch: lapak.CategoryBranchFinishedGoods,
					Children: []categoryNode{
						{Name: "Buku Catatan", Slug: "buku-catatan", Branch: lapak.CategoryBranchFinishedGoods},
						{Name: "Tempat Pensil", Slug: "tempat-pensil", Branch: lapak.CategoryBranchFinishedGoods},
					},
				},
				{
					Name:   "Dekorasi & Hiasan",
					Slug:   "dekorasi-hiasan",
					Branch: lapak.CategoryBranchFinishedGoods,
					Children: []categoryNode{
						{Name: "Hiasan Dinding", Slug: "hiasan-dinding", Branch: lapak.CategoryBranchFinishedGoods},
						{Name: "Pot Tanaman", Slug: "pot-tanaman", Branch: lapak.CategoryBranchFinishedGoods},
						{Name: "Lampu Hias Meja", Slug: "lampu-hias-meja", Branch: lapak.CategoryBranchFinishedGoods},
						{Name: "Celengan", Slug: "celengan", Branch: lapak.CategoryBranchFinishedGoods},
					},
				},
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


const (
	seedImagesDir   = "internal/database/seeding/images"
	uploadPhotosDir = "uploads/v1/photos"
	seedImageCount  = 10
)

var seedCoords = [][2]float64{
	{-6.2146, 106.8451},
	{-6.9175, 107.6191},
	{-6.2088, 106.8456},
	{-6.3025, 106.6523},
	{-6.1754, 106.8271},
	{-6.2152, 106.8500},
	{-6.2200, 106.8400},
	{-6.2100, 106.8600},
	{-6.2050, 106.8350},
	{-6.2250, 106.8550},
	{-6.1900, 106.8700},
	{-6.2300, 106.8300},
	{-6.2000, 106.8800},
	{-6.2400, 106.8200},
	{-6.1800, 106.8900},
	{-6.2500, 106.8100},
	{-6.1700, 106.9000},
	{-6.2600, 106.8000},
	{-6.1600, 106.9100},
	{-6.2700, 106.7900},
}

type productTemplate struct {
	Title       string
	Description string
}

var finishedGoodsTemplates = []productTemplate{
	{"Kursi Kayu Daur Ulang", "Kursi hasil daur ulang kayu bekas, kokoh dan siap pakai untuk rumah atau kafe."},
	{"Tas Rajut dari Limbah Plastik", "Tas rajut dari benang plastik daur ulang, ramah lingkungan dan tahan lama."},
	{"Pot Bunga dari Ban Bekas", "Pot bunga unik hasil kreasi dari ban bekas, cocok untuk taman minimalis."},
	{"Lampu Hias dari Botol Kaca", "Lampu hias dengan bahan dasar botol kaca bekas, memberi nuansa estetik di rumah."},
	{"Rak Buku Multifungsi", "Rak buku dari kayu palet daur ulang, kuat, ringan, dan tahan lama."},
	{"Vas Bunga Keramik Daur Ulang", "Vas bunga cantik dari bahan keramik daur ulang, cocok untuk dekorasi ruangan."},
	{"Tempat Pensil dari Kaleng Bekas", "Tempat pensil kreatif dari kaleng bekas yang dicat ulang dengan warna cerah."},
	{"Bantal Kursi dari Kain Perca", "Bantal kursi empuk dari kain perca berbagai motif yang dijahit rapi."},
}

var rawMaterialTemplates = []productTemplate{
	{"Kayu Palet Bekas", "Kayu palet bekas kondisi masih bagus, cocok untuk bahan baku furnitur daur ulang."},
	{"Botol Plastik PET Bersih", "Botol plastik PET bekas yang sudah dibersihkan, siap untuk didaur ulang."},
	{"Kain Perca Campuran", "Kumpulan kain perca berbagai warna dan tekstur, cocok untuk bahan kerajinan."},
	{"Kardus Bekas Layak Pakai", "Kardus bekas dalam kondisi kering dan masih kuat untuk dipakai kembali."},
	{"Besi Scrap Campuran", "Besi scrap sisa konstruksi, siap dilebur ulang atau dipakai untuk kerajinan logam."},
	{"Ban Bekas Mobil", "Ban bekas mobil kondisi masih utuh, cocok untuk kerajinan atau pot bunga."},
	{"Kaca Bekas Botol", "Botol kaca bekas berbagai ukuran, sudah dicuci bersih dan siap dipakai ulang."},
	{"Kertas Bekas Kantor", "Kertas bekas dari aktivitas kantor, cocok untuk didaur ulang menjadi kertas baru."},
}

func randomRange(min, max int) int {
	if min >= max {
		return min
	}
	return min + rand.Intn(max-min+1)
}

func randomCoord() (float64, float64) {
	c := seedCoords[rand.Intn(len(seedCoords))]
	return c[0], c[1]
}

func getLeafCategories(db *gorm.DB, branch string) ([]lapak.Category, error) {
	var categories []lapak.Category
	if err := db.Where("branch = ?", branch).Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("gagal mengambil kategori branch %s: %w", branch, err)
	}

	parentIDs := make(map[uuid.UUID]bool, len(categories))
	for _, c := range categories {
		if c.ParentID != nil {
			parentIDs[*c.ParentID] = true
		}
	}

	leaves := make([]lapak.Category, 0)
	for _, c := range categories {
		if !parentIDs[c.ID] {
			leaves = append(leaves, c)
		}
	}
	return leaves, nil
}

func ensureUploadDir() error {
	return os.MkdirAll(uploadPhotosDir, 0755)
}

func copyProductImage(index int, productID uuid.UUID) (string, error) {
	srcName := fmt.Sprintf("product_%d.jpg", (index%seedImageCount)+1)
	srcPath := filepath.Join(seedImagesDir, srcName)

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("gambar dummy %s tidak ditemukan: %w", srcPath, err)
	}
	defer srcFile.Close()

	destName := fmt.Sprintf("%s_%d.jpg", productID.String(), time.Now().UnixNano())
	destPath := filepath.Join(uploadPhotosDir, destName)

	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file tujuan %s: %w", destPath, err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return "", fmt.Errorf("gagal menyalin gambar: %w", err)
	}

	return "/" + destPath, nil
}

func SeedProducts(db *gorm.DB, users []user.User) error {
	var existingCount int64
	if err := db.Model(&lapak.Product{}).Count(&existingCount).Error; err != nil {
		return fmt.Errorf("gagal mengecek produk existing: %w", err)
	}
	if existingCount > 0 {
		log.Println("Produk sudah ada, skip seeding produk...")
		return nil
	}

	sellers := make([]user.User, 0)
	for _, u := range users {
		if u.Role == "seller" {
			sellers = append(sellers, u)
		}
	}
	if len(sellers) == 0 {
		return fmt.Errorf("tidak ada user dengan role seller, tidak bisa seeding produk")
	}

	if err := ensureUploadDir(); err != nil {
		return fmt.Errorf("gagal membuat folder upload %s: %w", uploadPhotosDir, err)
	}

	finishedLeaves, err := getLeafCategories(db, "FINISHED_GOODS")
	if err != nil {
		return err
	}
	rawLeaves, err := getLeafCategories(db, "RAW_MATERIAL")
	if err != nil {
		return err
	}
	if len(finishedLeaves) == 0 && len(rawLeaves) == 0 {
		return fmt.Errorf("tidak ditemukan kategori leaf, pastikan SeedCategories dijalankan lebih dulu")
	}

	imageIndex := 0
	totalProducts := 0
	totalImages := 0

	insertProduct := func(category lapak.Category, listingType string, price int, stock *int) error {
		seller := sellers[rand.Intn(len(sellers))]
		lat, lng := randomCoord()

		templates := finishedGoodsTemplates
		if category.Branch == "RAW_MATERIAL" {
			templates = rawMaterialTemplates
		}
		tpl := templates[rand.Intn(len(templates))]

		product := lapak.Product{
			ID:          uuid.New(),
			UserID:      seller.ID,
			Title:       fmt.Sprintf("%s - %s", tpl.Title, category.Name),
			Description: tpl.Description,
			CategoryID:  category.ID,
			ListingType: listingType,
			Price:       price,
			Status:      "AVAILABLE",
			Latitude:    lat,
			Longitude:   lng,
			Stock:       stock,
		}

		if err := db.Create(&product).Error; err != nil {
			return fmt.Errorf("gagal insert produk %s: %w", product.Title, err)
		}
		totalProducts++

		url, err := copyProductImage(imageIndex, product.ID)
		imageIndex++
		if err != nil {
			log.Printf("Peringatan: %v (produk %q dibuat tanpa gambar)\n", err, product.Title)
			return nil
		}

		image := lapak.ProductImage{
			ID:        uuid.New(),
			ProductID: product.ID,
			URL:       url,
		}
		if err := db.Create(&image).Error; err != nil {
			return fmt.Errorf("gagal insert product image: %w", err)
		}
		totalImages++
		return nil
	}

	for _, leaf := range finishedLeaves {
		for i := 0; i < 3; i++ {
			price := randomRange(10000, 50000)
			stock := randomRange(1, 10)
			if err := insertProduct(leaf, "DIJUAL", price, &stock); err != nil {
				return err
			}
		}
	}

	for _, leaf := range rawLeaves {
		for i := 0; i < 5; i++ {
			if i%2 == 0 {
				if err := insertProduct(leaf, "HIBAH", 0, nil); err != nil {
					return err
				}
			} else {
				price := randomRange(5000, 15000)
				if err := insertProduct(leaf, "JUAL_BORONGAN", price, nil); err != nil {
					return err
				}
			}
		}
	}

	log.Printf("Berhasil seeding %d produk (%d dengan gambar, %d leaf FINISHED_GOODS, %d leaf RAW_MATERIAL)\n",
		totalProducts, totalImages, len(finishedLeaves), len(rawLeaves))

	return nil
}
