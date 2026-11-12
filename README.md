# Growser

Минималистичный веб-браузер на Go с собственным движком рендеринга.

## Возможности

- HTTP-клиент с User-Agent и таймаутом
- HTML5-парсер (`golang.org/x/net/html`) в собственное DOM-дерево
- CSS-парсер с поддержкой селекторов (тег, класс, id), каскада и inline-стилей
- Блочная модель: `margin`, `padding`, `max-width`, `margin: auto`
- Свойства: `font-size`, `line-height`, `text-align`, `color`, `background`
- Единицы: `px`, `em`, `%`, `vh`, `vw`
- Раскладка с переносом текста по словам
- Inline-runs: отдельные Box для каждого `<a>` — точные клики
- Интерактивное окно на Ebiten: прокрутка, клики по ссылкам, hover
- История навигации (Alt+Left / Alt+Right, кнопки на панели)
- Адресная строка с вводом URL (Ctrl+L)
- Экспорт страницы в PNG (`-debug` для отладочных рамок)

## Установка

Требуется Go 1.21+.

```bash
go mod download
go run . -ui https://example.com