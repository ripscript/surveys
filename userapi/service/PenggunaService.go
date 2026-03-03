package service

import (
	"backend/siccore/pb"
	"backend/userapi/models"
	"backend/userapi/payloads"
	"backend/userapi/repository"
	"backend/userapi/utils"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type PenggunaService interface {
	Login(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	Register(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	Profile(usr models.JwtCustomClaims) (*pb.ProxyResponse, error)
	ConfirmProfile(req map[string]interface{}, usr models.JwtCustomClaims) (*pb.ProxyResponse, error)
	Logout(usr models.JwtCustomClaims) (*pb.ProxyResponse, error)
	EncryptInt() (*pb.ProxyResponse, error)
	ResetPassword(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ActiveUser(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Verification(usr models.JwtCustomClaims) (*pb.ProxyResponse, error)
	GetUserById(slug map[string]interface{}) (*pb.ProxyResponse, error)
	AddressDetail(usr models.JwtCustomClaims) (*pb.ProxyResponse, error)
	RequestResetPassword(req map[string]interface{}) (*pb.ProxyResponse, error)
	CompeleteResetPasswrod(req map[string]interface{}, usr models.JwtCustomClaims) (*pb.ProxyResponse, error)
	UpdateProfile(req map[string]interface{}, usr models.JwtCustomClaims) (*pb.ProxyResponse, error)
}

type penggunaService struct {
	penggunaRepo repository.PenggunaRepo
}

func NewPenggunaService(
	penggunaRepo repository.PenggunaRepo,

) PenggunaService {
	return &penggunaService{
		penggunaRepo,
	}
}

func (service *penggunaService) CompeleteResetPasswrod(req map[string]interface{}, usr models.JwtCustomClaims) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	userID := usr.ID
	password := req["password"].(string)
	rePassword := req["retypePassword"].(string)

	if password != rePassword {
		return utils.SendError(fmt.Errorf("Password Tidak Sama"), http.StatusBadRequest)
	}

	var dataPassword models.UpdatePasswordPenggune

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	dataPassword.ID = userID
	dataPassword.Password = string(hashedPassword)

	err = service.penggunaRepo.ResetPassword(dataPassword)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData("Password berhasil di ubah")
}

func (service *penggunaService) RequestResetPassword(req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	dataNew, err := service.penggunaRepo.GetUserByEmail(req["email"].(string))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.SendError(fmt.Errorf("Email tidak terdaftar"), http.StatusInternalServerError)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	claims := &models.JwtCustomClaims{
		ID:    dataNew.ID,
		Name:  dataNew.Name,
		Email: dataNew.Email,
		Role:  dataNew.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	baseUrl := os.Getenv("BASE_URL")

	subject := "Verifikasi Email"
	templatePath := getTemplatePath("confirmResetPassword.html")
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		return utils.SendError(errors.New("Gagal membuat token"), http.StatusInternalServerError)
	}
	url := baseUrl + "/forgot-password/" + tokenString

	Return := map[string]interface{}{
		"Nama": dataNew.Name,
		"Url":  url,
	}
	htmlBody, err := service.ParseHTMLTemplate(templatePath, Return)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	err = utils.SendEmailHtml(req["email"].(string), subject, htmlBody, true)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData("Silahkan cek email untuk melakukan reset password")
}

func (service *penggunaService) GetUserById(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	StrId := slug["id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	data, err := service.penggunaRepo.GetDetail(int(Id))
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	if data.AccountStatus == false {
		return utils.SendError(fmt.Errorf("akun tidak aktif"), http.StatusBadRequest)
	}
	return utils.SendData(data, "Berhasil mengambil data")
}

func (service *penggunaService) Verification(usr models.JwtCustomClaims) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	id := usr.ID

	err := service.penggunaRepo.Verification(int(id))
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData("Akun berhasil di verifikasi")
}

func (service *penggunaService) Login(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.LoginPayload

	// mengambil payload
	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	// Validasi input pengguna
	if payload.Email == "" || payload.Password == "" {
		err := errors.New("Email dan password diperlukan")
		return utils.SendError(err, http.StatusBadRequest)
	}

	// Validasi Memeriksa kredensial pengguna
	storedUser, err := service.penggunaRepo.ValidasiCredential(payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	if storedUser.AccountStatus == false {
		return utils.SendError(fmt.Errorf("akun belum aktif, silahkan buka pesan verifikasi"), http.StatusUnauthorized)
	}

	// Menetapkan klaim kustom
	claims := &models.JwtCustomClaims{
		ID:    int64(storedUser.ID),
		Name:  storedUser.Name,
		Email: storedUser.Email,
		Role:  storedUser.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	// Membuat token dengan klaim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Menghasilkan token terenkripsi
	encryptedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		err := errors.New("Gagal membuat token")
		return utils.SendError(err, http.StatusBadRequest)
	}

	// Update LastLog ke waktu sekarang
	storedUser.LastLogin = time.Now().In(utils.TimeNow())
	service.penggunaRepo.EditLastLog(storedUser)

	// return utils.SendData(nil, "OTP telah dikirim")
	return utils.SendData(encryptedToken, "Login Berhasil")
}

func (service *penggunaService) Register(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.CreatePengguna

	payload.Role = "customer"
	payload.PhoneNumber = "-"
	payload.Name = "-"

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	if payload.Email == "" || payload.Password == "" {
		err := errors.New("Email dan password diperlukan")
		return utils.SendError(err, http.StatusBadRequest)
	}

	if payload.Password != payload.ConfirmPassword {
		err := errors.New("Konfirmasi password tidak valid")
		return utils.SendError(err, http.StatusBadRequest)
	}

	var data models.Users

	err = utils.DynamicBind(payload, &data)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	data.AccountStatus = true

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	data.Password = string(hashedPassword)

	err = service.penggunaRepo.ValidasiPengguna(data, "add", "register")
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	pengguna, errs := service.penggunaRepo.Add(data)
	if errs != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	claims := &models.JwtCustomClaims{
		ID:    int64(pengguna.ID),
		Name:  pengguna.Name,
		Email: pengguna.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	encryptedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		err := errors.New("Gagal membuat token")
		return utils.SendError(err, http.StatusBadRequest)
	}

	return utils.SendData(encryptedToken, "Berhasil menyimpan data")
}

func (s *penggunaService) sendVerifEmail(tokenString string, name string, email string) error {
	baseUrl := os.Getenv("BASE_URL")

	subject := "Verifikasi Email"
	templatePath := getTemplatePath("EmailVerificationRegist.html")
	url := baseUrl + "/verify-account/" + tokenString

	Return := map[string]interface{}{
		"Nama":    name,
		"Url":     url,
		"BaseUrl": baseUrl + "/api",
	}
	htmlBody, err := s.ParseHTMLTemplate(templatePath, Return)
	if err != nil {
		return err
	}
	err = utils.SendEmailHtml(email, subject, htmlBody, true)
	if err != nil {
		return err
	}

	return nil
}

func (service *penggunaService) Profile(usr models.JwtCustomClaims) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	data, err := service.penggunaRepo.GetProfile(usr)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	return utils.SendData(data, "Berhasil mengambil data")
}

func (service *penggunaService) ConfirmProfile(req map[string]interface{}, usr models.JwtCustomClaims) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var ComplatedProfile payloads.CompleteProfile
	ComplatedProfile.ID = usr.ID

	err := utils.DynamicBind(req, &ComplatedProfile)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var UserBank payloads.UserBank

	if ComplatedProfile.Jastiper {
		EncBankId := req["bankId"].(string)
		DecBankId, err := utils.DecryptInt(EncBankId)
		if err != nil {
			return utils.SendError(fmt.Errorf("gagal mendapatkan bank ID"), http.StatusBadRequest)
		}

		UserBank.UserId = usr.ID
		UserBank.BankId = int64(DecBankId)

		err = utils.DynamicBind(req, &UserBank)
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
	}
	var dataCompleated models.CompleteProfile

	if ComplatedProfile.Jastiper {
		dataCompleated.Role = "jastiper"
		status := "pending"
		dataCompleated.JastiperStatus = &status
	} else {
		dataCompleated.Role = "customer"
	}

	err = utils.DynamicBind(ComplatedProfile, &dataCompleated)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	Slug := toCamelCase(dataCompleated.Name)

	dataCompleated.Slug = Slug

	address, ok := req["address"].(map[string]interface{})
	if !ok {
		return utils.SendError(fmt.Errorf("format address tidak valid"), http.StatusBadRequest)
	}
	cityString, _ := address["city"].(int)
	provinceString, _ := address["province"].(int)
	detail, _ := address["detail"].(string)

	cityInt := cityString

	provinceInt := provinceString

	dataCompleated.CityID = int64(cityInt)
	dataCompleated.ProvinceID = int64(provinceInt)
	dataCompleated.Address = detail

	err = service.penggunaRepo.CompleteProfile(dataCompleated)
	if err != nil {
		return utils.SendError(fmt.Errorf("gagal memperbarui Profil"+err.Error()), http.StatusInternalServerError)
	}

	if ComplatedProfile.Jastiper {
		var dataUserBank models.UserBank
		err = utils.DynamicBind(UserBank, &dataUserBank)
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}

		err = service.penggunaRepo.UserBank(dataUserBank)
		if err != nil {
			return utils.SendError(fmt.Errorf("gagal memperbarui Profil"+err.Error()), http.StatusInternalServerError)
		}
	}

	claims := &models.JwtCustomClaims{
		ID:    int64(dataCompleated.ID),
		Name:  dataCompleated.Name,
		Email: usr.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	encryptedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		err := errors.New("Gagal membuat token")
		return utils.SendError(err, http.StatusBadRequest)
	}

	err = service.sendVerifEmail(encryptedToken, dataCompleated.Name, usr.Email)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData("Data Berhasil di selesaikan, Silahkan cek email anda untuk melakukan verifikasi akun")
}

func (service *penggunaService) Logout(usr models.JwtCustomClaims) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	return utils.SendData(nil, "Berhasil logout")
}

func (service *penggunaService) EncryptInt() (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	encrypt, err := utils.EncryptInt(1)
	if err != nil {
		return utils.SendError(fmt.Errorf("gagal mendapatkan encrypt data"), http.StatusInternalServerError)
	}

	return utils.SendData(encrypt)
}

func (service *penggunaService) ResetPassword(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	Encid := slug["id"].(string)
	Id, err := utils.DecryptInt(Encid)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	data, err := service.penggunaRepo.GetDetail(Id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if data.AccountStatus == false {
		return utils.SendError(fmt.Errorf("Akun tidak aktif"), http.StatusBadRequest)
	}

	password, err := service.GeneratePassword()
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var model models.UpdatePasswordPenggune

	model.ID = data.ID
	model.Password = string(hashedPassword)
	model.UpdatedAt = time.Now()
	model.UpdatedBy = usr.ID

	err = service.penggunaRepo.ResetPassword(model)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	code := "Reset Password"
	err = service.SendEmail(data.Email, data.Name, password, code)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData("Berhasil mengirim email reset password")
}

func (service *penggunaService) GeneratePassword() (string, error) {
	lettersLower := "abcdefghijklmnopqrstuvwxyz"
	lettersUpper := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"
	symbols := "!@#$%^&*()_+-=[]{}\\|;:\"<>/?"

	passwordLength := 8
	if passwordLength < 4 {
		return "", fmt.Errorf("password length must be at least 4")
	}
	upper := lettersUpper[rand.Intn(len(lettersUpper))]
	num := numbers[rand.Intn(len(numbers))]
	symbol := symbols[rand.Intn(len(symbols))]
	remainingLength := passwordLength - 3
	passwordChars := []byte{upper, num, symbol}
	for i := 0; i < remainingLength; i++ {
		passwordChars = append(passwordChars, lettersLower[rand.Intn(len(lettersLower))])
	}
	rand.Shuffle(len(passwordChars), func(i, j int) {
		passwordChars[i], passwordChars[j] = passwordChars[j], passwordChars[i]
	})

	return string(passwordChars), nil
}

func (s *penggunaService) SendEmail(email string, name string, data string, code string) error {
	subject := "Reset Password Akun Jastiper"
	templatePath := getTemplatePath("emailPassword.html")
	Return := map[string]interface{}{
		"Nama":     name,
		"Password": data,
	}
	htmlBody, err := s.ParseHTMLTemplate(templatePath, Return)
	if err != nil {
		return err
	}
	err = utils.SendEmailHtml(email, subject, htmlBody, true)
	if err != nil {
		return err
	}

	return nil
}

func (s *penggunaService) ParseHTMLTemplate(templatePath string, data map[string]interface{}) (string, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("gagal membaca file template: %v", err)
	}
	htmlStr := string(content)

	tmpl, err := template.New("emailTemplate").Parse(htmlStr)
	if err != nil {
		return "", fmt.Errorf("gagal parsing template: %v", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("gagal mengeksekusi template: %v", err)
	}

	return buf.String(), nil
}

func getTemplatePath(filename string) string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	basepath = filepath.Dir(basepath)
	return filepath.Join(basepath, "template", filename)
}

func (s *penggunaService) ActiveUser(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	Encid := slug["id"].(string)
	Id, err := utils.DecryptInt(Encid)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	data, err := s.penggunaRepo.GetDetail(Id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	var model models.UpdateActiveUser

	model.ID = data.ID

	if data.AccountStatus == true {
		model.AccountStatus = false
	} else {
		model.AccountStatus = true
	}

	model.UpdatedAt = time.Now()
	model.UpdatedBy = usr.ID

	err = s.penggunaRepo.ActiveUser(model)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData("Berhasil merubah status akun")
}

func (s *penggunaService) AddressDetail(usr models.JwtCustomClaims) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	userId, err := utils.ToInt64(usr.ID)
	if err != nil {
		return utils.SendError(fmt.Errorf("User id tidak ditemukan"), http.StatusBadRequest)
	}

	data, err := s.penggunaRepo.AddressDetail(int(userId))
	if err != nil {
		return utils.SendData(nil)
	}

	return utils.SendData(data)
}

func toCamelCase(input string) string {
	words := strings.Fields(input)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, "")
}

func (s *penggunaService) UpdateProfile(req map[string]interface{}, usr models.JwtCustomClaims) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	userID := usr.ID
	var payload interface{}
	data := models.UpdateProfile{}

	if usr.Role == "jastiper" {
		val := payloads.UpdateProfileJastiper{}
		err := utils.DynamicBind(req, &val)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		val.ID = int(userID)
		payload = val

		err = utils.DynamicBind(payload, &data)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		err = s.penggunaRepo.ValidasiPenggunaUpdate(data)
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
		BankId, err := utils.DecryptInt(req["bank_id"].(string))
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		UserBank := models.UserBank{}
		UserBank.UserID = userID
		UserBank.BankID = int64(BankId)
		UserBank.AccountNumber = val.NoRek
		err = s.penggunaRepo.UpdateBankProfile(UserBank)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

	} else {
		val := payloads.UpdateProfileUmum{}
		err := utils.DynamicBind(req, &val)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		val.ID = int(userID)
		payload = val
		err = utils.DynamicBind(payload, &data)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		err = s.penggunaRepo.ValidasiPenggunaUpdate(data)
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
	}

	err := s.penggunaRepo.UpdateProfile(data)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData("Data Berhasil Di Ubah")
}
