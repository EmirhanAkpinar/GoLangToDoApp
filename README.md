
# GoLangToDoApp 

Bu basit ToDo API Go programlama dili kullanılarak yazılmıştır. Go ile daha önceden fazla çalışmadığım için biraz amatörce olmuş olabilir. 

## Kurulum

Go kurulu olmalıdır. [Go Kurulum](https://golang.org/doc/install)

Projeyi klonlayın:

    git clone https://github.com/EmirhanAkpinar/GoLangToDoApp.git

Proje dizinine gidin:

    cd GoLangToDoApp

Uygulamayı çalıştırın:

    go run main.go




## API Kullanımı



#### Giriş Yap

```http
  GET /login
```

| Parametre | Tip     | Açıklama                |
| :-------- | :------- | :------------------------- |
| `Username` | `string` | **Gerekli**. Kullanıcı Adı |
| `Password` | `string` | **Gerekli**. Şifre |

---
#### Tüm Listeleri Getir

```http
  GET /list
```
---
#### Belirtilen Listeyi ve Görevleri Getir

```http
  GET /list/${id}
```

| Parametre | Tip     | Açıklama                |
| :-------- | :------- | :------------------------- |
| `id` | `int` | **Gerekli**. Listenin id değeri. |

---
#### Belirtilen Listenin Başlığını Düzenle

```http
  GET /list/${id}/update
```

| Parametre | Tip     | Açıklama                |
| :-------- | :------- | :------------------------- |
| `id` | `int` | **Gerekli**. Listenin id değeri. |
| `Title` | `string` | **Gerekli**. Listenin başlığı. |

---
#### Belirtilen Listede yeni görev ekle

```http
  GET /list/${id}/create
```

| Parametre | Tip     | Açıklama                       |
| :-------- | :------- | :-------------------------------- |
| `id` | `int` | **Gerekli**. Listenin id değeri. |
| `Task` | `string` | **Gerekli**. Görevin açıklaması. |

---
#### Belirtilen Listede Belirtilen Görevi Getir

```http
  GET /list/${id}/task/${taskid}
```

| Parametre | Tip     | Açıklama                       |
| :-------- | :------- | :-------------------------------- |
| `id` | `int` | **Gerekli**. Listenin id değeri. |
| `taskid` | `int` | **Gerekli**. Taskın id değeri. |

---
#### Belirtilen Listede Belirtilen Görevi Düzenle

```http
  GET /list/${id}/task/${taskid}/update
```

| Parametre | Tip     | Açıklama                       |
| :-------- | :------- | :-------------------------------- |
| `id` | `int` | **Gerekli**. Listenin id değeri. |
| `taskid` | `int` | **Gerekli**. Taskın id değeri. |
| `Task` | `string` | **Gerekli**. Görevin açıklaması. |

---
#### Belirtilen Listede Belirtilen Görevi Tamamlanmış Olarak İşaretle

```http
  GET /list/${id}/task/${taskid}/complete
```

| Parametre | Tip     | Açıklama                       |
| :-------- | :------- | :-------------------------------- |
| `id` | `int` | **Gerekli**. Listenin id değeri. |
| `taskid` | `int` | **Gerekli**. Taskın id değeri. |

---
#### Belirtilen Listede Belirtilen Görevi Sil

```http
  GET /list/${id}/task/${taskid}/delete
```

| Parametre | Tip     | Açıklama                       |
| :-------- | :------- | :-------------------------------- |
| `id` | `int` | **Gerekli**. Listenin id değeri. |
| `taskid` | `int` | **Gerekli**. Taskın id değeri. |

  