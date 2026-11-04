# 🛠️ ProxyGuide Core Builder

Автоматический конвейер сборки чистых нативных ядер **Xray-core** (`libv2ray.aar`) и **Mihomo** (`libmihomo.so`) для Android-клиента ProxyGuide.

---

### 🚀 Что собирается в этом репозитории:

1. **`libv2ray.aar` (Xray-core)**
   * **Протоколы:** VLESS (Reality, Vision), VMess, Trojan, Shadowsocks-2022 (SIP022, UoT).
   * **Транспорты:** XHTTP (SplitHTTP с `x_padding_bytes`), WebSocket, gRPC, HTTPUpgrade, TCP.
   * **Функции:** Фрагментация ClientHello против ТСПУ/DPI, FakeDNS, замеры задержки (`measureOutboundDelay`), счётчики трафика.
   * **Оптимизация:** Флаги `-ldflags="-s -w -buildid="` и `-trimpath` (срезает 40% лишнего веса).

2. **`libmihomo.so` (Mihomo MetaCubeX)**
   * **Протоколы:** AmneziaWG (рандомизация заголовков `h1`-`h4`, мусорные пакеты `jc`, `jmin`, `jmax`, `s1`, `s2`), Hysteria 2 (QUIC, port-hopping).
   * **Стек:** Встроенный TCP/IP стек **Google gVisor** (`-tags "with_gvisor"`).
   * **Архитектуры:** `arm64-v8a`, `armeabi-v7a`, `x86_64`.

---

### 📦 Как запустить сборку:

1. Откройте вкладку **[Actions](https://github.com/XZ-LITE/proxyguide-core-builder/actions)** в вашем репозитории.
2. В левой колонке выберите воркфлоу **«Build ProxyGuide Cores (Xray & Mihomo)»**.
3. Нажмите кнопку **«Run workflow»** ➔ **«Run workflow»**.
4. Сборка займет около 3 минут на серверах GitHub.
5. Во вкладке **Releases** появится готовый релиз с файлом **`ProxyGuide-Cores.zip`**!

---

### 📥 Как обновить бинарники в ProxyGuide:

1. Распакуйте скачанный архив `ProxyGuide-Cores.zip`.
2. Замените файл `app/libs/libv2ray.aar` в проекте ProxyGuide.
3. Скопируйте папки `arm64-v8a` и `armeabi-v7a` в `app/src/main/jniLibs/`.
4. Сделайте **Rebuild Project** в Android Studio.
