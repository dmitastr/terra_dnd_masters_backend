FROM python:3.12-slim

# Не создаём .pyc файлы и не буферизуем вывод — удобнее смотреть логи
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

# Сначала копируем только requirements — чтобы кешировался слой с зависимостями
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Копируем весь проект целиком
COPY . .

# Запускаем от непривилегированного пользователя — хорошая практика
RUN mkdir -p /app/db && useradd --create-home appuser && chown -R appuser:appuser /app
USER appuser

CMD ["python", "main.py"]