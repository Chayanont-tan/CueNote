# Stage 1: Build binary
FROM golang:1.25-alpine AS builder

WORKDIR /app

# ก็อปปี้ไฟล์ dependency และดาวน์โหลด modules ล่วงหน้า (ใช้ cache ได้ถ้า dependencies ไม่เปลี่ยน)
COPY go.mod go.sum ./
RUN go mod download

# ก็อปปี้ source code ทั้งหมดเข้ามา
COPY . .

# คอมไพล์ Binary (CGO_ENABLED=0 เพื่อให้เป็น Pure Go Static Binary รันบน Alpine ได้ไม่มีปัญหา)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/api ./cmd/api/main.go

# Stage 2: Minimal runtime
FROM alpine:latest

# ติดตั้ง Root Certificates (จำเป็นสำหรับการยิง HTTPS ไปยังภายนอก เช่น OpenAI API) และ Timezone
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# ก็อปปี้เฉพาะ Binary และโฟลเดอร์ที่จำเป็นต้องใช้ตอน runtime จาก Stage 1
COPY --from=builder /app/api /app/api
COPY --from=builder /app/assets /app/assets
COPY --from=builder /app/db/migrations /app/db/migrations

EXPOSE 8080

CMD ["/app/api"]