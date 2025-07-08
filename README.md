# Authentication API Backend

Backend API untuk sistem autentikasi menggunakan Go dengan JWT token.

## Fitur

- **Register User**: Endpoint untuk mendaftarkan user baru dengan auto-generated password
- **Login**: Endpoint untuk login dan mendapatkan JWT token
- **Profile**: Endpoint untuk mengambil data user yang dilindungi JWT
- **Environment Variables**: Konfigurasi yang aman menggunakan file .env
- **Error Handling**: Penanganan error yang lengkap
- **JWT Middleware**: Validasi token otomatis

## Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/health` | Health check | No |
| POST | `/api/register` | Register user baru | No |
| POST | `/api/login` | Login dan dapatkan token | No |
| GET | `/api/profile` | Ambil data profile user | Yes (JWT) |

## Setup

### 1. Clone Repository

```bash
git clone <repository-url>
cd auth-api
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Setup Environment Variables

Copy file `.env.example` ke `.env`:

```bash
cp .env.example .env
```

Edit file `.env` dan sesuaikan konfigurasi:

```bash
# JWT Secret Key - GANTI DENGAN SECRET KEY YANG KUAT DI PRODUCTION!
JWT_SECRET_KEY=your-super-secret-jwt-key-here

# Server Port
PORT=8080
```

> **⚠️ PENTING**: Jangan pernah commit file `.env` ke repository. Gunakan secret key yang kuat dan unik di production.

### 4. Jalankan Server

```bash
go run main.go
```

Server akan berjalan di `http://localhost:8080`

## Usage Examples

### 1. Register User

```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "nik": "1234567890123456",
    "role": "admin"
  }'
```

**Response:**
```json
{
  "id": 1,
  "nik": "1234567890123456",
  "role": "admin", 
  "password": "Abc123",
  "message": "User berhasil didaftarkan"
}
```

### 2. Login

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "nik": "1234567890123456",
    "password": "Abc123"
  }'
```

**Response:**
```json
{
  "id": 1,
  "nik": "1234567890123456",
  "role": "admin",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Login berhasil"
}
```

### 3. Get Profile

```bash
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "id": 1,
  "nik": "1234567890123456", 
  "role": "admin",
  "message": "Data profile berhasil diambil"
}
```

## Project Structure

```
auth-api/
├── main.go              # Main application file
├── go.mod               # Go module dependencies
├── .env                 # Environment variables (tidak di-commit)
├── .env.example         # Template environment variables
├── .gitignore           # Git ignore configuration
└── README.md            # Documentation
```

## Security Features

- **JWT Token**: Menggunakan JWT untuk autentikasi dengan expiry time
- **Environment Variables**: Secret key disimpan dalam environment variables
- **Input Validation**: Validasi input pada semua endpoint
- **Error Handling**: Error messages yang informatif tanpa expose sensitive data
- **Password Generation**: Auto-generated password dengan kombinasi huruf dan angka

## Production Deployment

Untuk deployment production:

1. **Generate Strong JWT Secret**:
   ```bash
   openssl rand -base64 32
   ```

2. **Set Environment Variables**:
   ```bash
   export JWT_SECRET_KEY="your-generated-secret-key"
   export PORT=8080
   ```

3. **Build Binary**:
   ```bash
   go build -o auth-api main.go
   ```

4. **Run Binary**:
   ```bash
   ./auth-api
   ```

## Dependencies

- [gorilla/mux](https://github.com/gorilla/mux) - HTTP router
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) - JWT implementation
- [joho/godotenv](https://github.com/joho/godotenv) - Environment variable loader

## Development

### Running Tests

```bash
go test ./...
```

### Code Format

```bash
go fmt ./...
```

### Build

```bash
go build -o auth-api main.go
```

## Contributing

1. Fork repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## License

MIT License