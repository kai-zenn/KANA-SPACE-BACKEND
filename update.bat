@echo off
echo Checking for updates...
echo [1] Menarik perubahan terbaru dari GitHub...
git pull origin main
if %errorlevel% neq 0 (
    echo ❌ Gagal pull dari GitHub!
    echo    Cek koneksi internet atau remote URL.
    pause
    exit /b 1
)
echo Pull berhasil!
echo.

echo [2] Menghentikan container lama...
docker-compose down
echo.

echo [3] Membangun ulang image...
docker-compose build --no-cache
echo.

echo [4] Menjalankan container...
docker-compose up -d
echo.

echo [5] Menjalankan migrasi database...
docker-compose exec app ./main migrate
echo.

echo [6] Cek kesehatan aplikasi...
timeout /t 3 /nobreak >nul
curl -f http://localhost:9090/health >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ Health check OK! Aplikasi berjalan.
) else (
    echo ⚠️ Health check gagal, tapi aplikasi mungkin tetap berjalan.
    echo    Cek log: docker-compose logs app
)

echo.
echo ========================================
echo ✅ UPDATE SELESAI!
echo    Aplikasi berjalan di http://localhost:9090
echo ========================================
pause
