# Название выходного архива
ARCHIVE_NAME=project.zip

# Файл игнор-листа
IGNORE_FILE=.zipignore

# Список файлов для архивации, исключая игнорируемые
FILES_TO_ARCHIVE=$(shell find . -type f | grep -v -F -f $(IGNORE_FILE))

# Валидация проекта
validate:
	@echo "Running go vet..."
	@go vet ./...
	@echo "Running tests..."
	@go test ./...

# Архивация проекта
zip: validate
	@echo "Creating archive..."
	@zip -r $(ARCHIVE_NAME) $(FILES_TO_ARCHIVE)
	@echo "Archive $(ARCHIVE_NAME) created."

# Очистка архива
clean:
	@rm -f $(ARCHIVE_NAME)
	@echo "Removed $(ARCHIVE_NAME)."

.PHONY: validate zip clean