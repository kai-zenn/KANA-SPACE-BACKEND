package seeding

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"KANA-SPACE-BACKEND/internal/modules/space"
	"KANA-SPACE-BACKEND/internal/modules/user"
)

var sellerPostContents = []string{
	"Saya punya banyak kayu palet bekas kondisi bagus, cocok banget buat yang mau bikin furnitur daur ulang. Yuk mampir ke lapak saya!",
	"Lagi banyak stok kain perca warna-warni nih, sayang kalau dibuang. Bisa dipakai buat kerajinan atau bantal custom.",
	"Tersedia besi scrap sisa proyek konstruksi, kondisi masih layak pakai untuk dilebur ulang atau dibuat kerajinan logam.",
	"Ada beberapa ban bekas mobil yang masih bagus, cocok dijadikan pot bunga atau ayunan taman, harga bisa nego.",
	"Kardus bekas melimpah di gudang saya, kondisinya kering dan kuat, siap didaur ulang jadi barang baru.",
	"Botol kaca bekas berbagai ukuran sudah saya bersihkan, cocok untuk dijadikan lampu hias atau vas bunga yang unik.",
}

var buyerPostContents = []string{
	"Lagi cari material kayu bekas buat proyek DIY di rumah, ada yang punya stok lebih? Boleh info harga dan lokasinya.",
	"Butuh kain perca dalam jumlah banyak untuk workshop menjahit komunitas, semoga ada yang bisa bantu ya.",
	"Mencari botol plastik bekas untuk didaur ulang jadi kerajinan tangan, kalau ada yang jual borongan boleh chat saya.",
	"Sedang cari besi scrap untuk proyek seni instalasi, kalau ada yang jual di sekitar Jakarta boleh info ya.",
	"Butuh beberapa ban bekas untuk dijadikan pot tanaman di halaman rumah, semoga ada yang bersedia menghibahkan.",
	"Cari kardus bekas dalam kondisi baik untuk kebutuhan packing usaha kecil saya, kalau ada yang jual boleh kontak.",
}

// SeedPosts membuat 1-3 postingan acak untuk setiap user (seller maupun buyer).
// Seller mendapat tag "JualMaterial", buyer mendapat tag "CariMaterial", dan
// koordinat postingan diambil acak dari daftar seedCoords (didefinisikan di
// lapak_seed.go) agar konsisten dengan sebaran koordinat produk.
func SeedPosts(db *gorm.DB, users []user.User) error {
	var existingCount int64
	if err := db.Model(&space.Post{}).Count(&existingCount).Error; err != nil {
		return fmt.Errorf("gagal mengecek post existing: %w", err)
	}
	if existingCount > 0 {
		log.Println("Postingan sudah ada, skip seeding post...")
		return nil
	}

	totalPosts := 0
	for _, u := range users {
		postCount := randomRange(1, 3)

		for i := 0; i < postCount; i++ {
			var content, tag string
			if u.Role == "seller" {
				content = sellerPostContents[rand.Intn(len(sellerPostContents))]
				tag = "JualMaterial"
			} else {
				content = buyerPostContents[rand.Intn(len(buyerPostContents))]
				tag = "CariMaterial"
			}

			lat, lng := randomCoord()

			post := space.Post{
				ID:        uuid.New(),
				UserID:    u.ID,
				Content:   content,
				Tag:       tag,
				Latitude:  &lat,
				Longitude: &lng,
			}

			if err := db.Create(&post).Error; err != nil {
				return fmt.Errorf("gagal insert post untuk user %s: %w", u.Username, err)
			}
			totalPosts++
		}
	}

	log.Printf("Berhasil seeding %d postingan dari %d user\n", totalPosts, len(users))
	return nil
}
