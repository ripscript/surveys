## 1. Buat Virtual Environment

Masuk ke folder project, lalu buat env baru (umum dinamai `venv` atau `.venv`):

```bash
python3 -m venv venv
```

## 2. Aktifkan Virtual Environment

```bash
source venv/bin/activate
```

Setelah aktif, prompt terminal akan menampilkan `(venv)` di depan, contoh:

```
(venv) user@host:~/project$
```

## 3. Install Dependency dari requirements.txt

Pastikan file `requirements.txt` ada di folder project, lalu jalankan:

```bash
pip install -r requirements.txt
```

## 4. Jalankan Aplikasi

```bash
python main.py
```

## 6. Menonaktifkan Virtual Environment

Setelah selesai kerja:

```bash
deactivate
```

---

### Ringkasan Perintah Cepat

```bash
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python main.py
```