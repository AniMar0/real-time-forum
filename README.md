# Real-time Forum

منتدى بسيط يدعم المنشورات، التعليقات، المحادثات الفورية عبر WebSocket، والإشعارات غير المقروءة.

## المتطلبات

- Go 1.25 أو أحدث
- SQLite
- Docker و Docker Compose (اختياري)

## التشغيل محليًا

```bash
go run .
```

افتح:

```text
http://localhost:8080
```

يتم إنشاء قاعدة البيانات وتشغيل الـ migrations تلقائيًا عند بدء التطبيق.

## التشغيل باستخدام Docker

```bash
docker compose up --build
```

ثم افتح:

```text
http://localhost:8080
```

يتم حفظ SQLite داخل Docker volume باسم `forum-data` حتى تبقى البيانات عند إعادة إنشاء الحاوية.

لإيقاف الحاوية:

```bash
docker compose down
```

## متغيرات البيئة

| المتغير | القيمة الافتراضية | الوصف |
|---|---|---|
| `FORUM_HTTP_ADDRESS` | `:8080` | عنوان ومنفذ HTTP |
| `FORUM_DATABASE_PATH` | `database/forum.db` | مسار SQLite |
| `FORUM_STATIC_PATH` | `static` | مسار ملفات الواجهة |
| `FORUM_ENV` | `development` | بيئة التشغيل؛ production يفعّل Secure cookies |
| `FORUM_WS_ORIGINS` | localhost origins | Origins المسموح بها للـ WebSocket، مفصولة بفواصل |

مثال:

```bash
FORUM_HTTP_ADDRESS=:9090 \
FORUM_DATABASE_PATH=database/forum.db \
FORUM_WS_ORIGINS=http://localhost:9090 \
go run .
```

## الوظائف الحالية

- إنشاء حساب وتسجيل الدخول باستخدام البريد أو nickname.
- جلسات دخول محفوظة في SQLite.
- إنشاء وعرض المنشورات.
- إضافة وعرض التعليقات.
- محادثة مباشرة عبر WebSocket.
- حفظ الرسائل والإشعارات في معاملة قاعدة بيانات واحدة.
- تحميل سجل المحادثة السابق.
- إغلاق الدردشة ودعم العرض المتجاوب على الهاتف.

## المسارات الرئيسية

| المسار | الطريقة | الاستخدام |
|---|---|---|
| `/register` | POST | إنشاء حساب |
| `/login` | POST | تسجيل الدخول |
| `/logout` | POST | تسجيل الخروج |
| `/logged` | POST | التحقق من الجلسة |
| `/posts` | GET | جلب المنشورات |
| `/createPost` | POST | إنشاء منشور |
| `/comments` | GET | جلب التعليقات |
| `/createComment` | POST | إنشاء تعليق |
| `/messages` | POST | جلب سجل المحادثة |
| `/notifications` | GET | جلب الإشعارات |
| `/notifications/mark-read` | POST | تحديد الإشعار كمقروء |
| `/ws` | WebSocket | المحادثة والحالة والكتابة |

## بنية المشروع

```text
.
├── backend/
│   ├── account/       # الجلسات وبيانات الحساب
│   ├── chat/          # الرسائل وسجل المحادثة
│   ├── forum/         # المنشورات والتعليقات
│   ├── notification/  # الإشعارات
│   └── migrations/    # ترحيلات SQLite
├── static/            # HTML/CSS/JavaScript للواجهة
├── database/          # قاعدة SQLite المحلية
├── Dockerfile
└── docker-compose.yml
```

## الاختبارات

```bash
go test ./...
go test -race ./...
go vet ./...
go build .
```

## ملاحظات قاعدة البيانات

قاعدة البيانات المحلية `database/forum.db` مخصصة للتشغيل المحلي. في Docker يتم استخدام volume منفصل. لا تحذف ملف قاعدة البيانات أثناء تشغيل التطبيق.
