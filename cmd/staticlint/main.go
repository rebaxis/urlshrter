// Package main реализует статический анализатор кода staticlint.
//
// staticlint - это комплексный инструмент статического анализа кода Go,
// объединяющий множество проверок в одном multichecker.
//
// # Установка
//
// Соберите инструмент из исходного кода:
//
//	go build -o staticlint ./cmd/staticlint
//
// Или установите напрямую:
//
//	go install github.com/rebaxis/urlshrter/cmd/staticlint@latest
//
// # Использование
//
// Базовое использование - проверка всех пакетов в текущей директории:
//
//	staticlint ./...
//
// Проверка конкретного пакета:
//
//	staticlint ./internal/service
//
// Проверка конкретного файла:
//
//	staticlint main.go
//
// # Включение/отключение анализаторов
//
// По умолчанию все анализаторы включены. Вы можете включить только определенные:
//
//	staticlint -osexit -nakedret ./...
//
// Или отключить все и включить нужные:
//
//	staticlint -checks="" -nakedret -osexit ./...
//
// Посмотреть список всех доступных анализаторов:
//
//	staticlint -help
//
// # Примеры использования
//
// Проверка на naked returns:
//
//	staticlint -nakedret ./...
//
// Проверка на использование os.Exit в main:
//
//	staticlint -osexit ./cmd/...
//
// Запуск всех SA анализаторов staticcheck:
//
//	staticlint -checks='SA*' ./...
//
// Комбинация нескольких проверок:
//
//	staticlint -osexit -nakedret -printf ./...
//
// # Включенные анализаторы
//
// Стандартные анализаторы Go (49):
//   - appends, asmdecl, assign, atomic, atomicalign, bools, buildssa,
//   - buildtag, cgocall, composite, copylock, ctrlflow, deepequalerrors,
//   - defers, directive, errorsas, fieldalignment, findcall, framepointer,
//   - httpmux, httpresponse, ifaceassert, inspect, loopclosure, lostcancel,
//   - nilfunc, nilness, printf, reflectvaluecompare, shadow, shift,
//   - sigchanyzer, slog, sortslice, stdmethods, stdversion, stringintconv,
//   - structtag, testinggoroutine, tests, timeformat, unmarshal, unreachable,
//   - unsafeptr, unusedresult, unusedwrite, usesgenerics
//
// Дополнительные анализаторы:
//   - nakedret: поиск naked returns в длинных функциях (>5 строк)
//   - osexit: проверка os.Exit в main функции main пакета
//
// Анализаторы staticcheck.io:
//   - SA (Static Analysis): 95 анализаторов для обнаружения багов и проблем
//   - ST (Style): 1 анализатор стиля кода
//   - S (Simple): 1 анализатор упрощения кода
//   - QF (Quickfix): 1 анализатор быстрых исправлений
//
// Итого: 178 анализаторов
//
// # Конфигурация
//
// Некоторые анализаторы поддерживают дополнительные флаги:
//
// nakedret:
//
//	Настраивается в коде (MaxLength: 5 строк)
//
// osexit:
//
//	Проверяет только main функцию в main пакете
//
// # Выходной код
//
// staticlint возвращает:
//   - 0: если проблем не найдено
//   - 3: если найдены проблемы в коде
//   - другое: при ошибке выполнения

package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"

	"github.com/alexkohler/nakedret/v2"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// main инициализирует и запускает multichecker со всеми настроенными анализаторами.
// Использует multichecker.Main для обработки флагов командной строки и выполнения проверок.
func main() {
	// Настройка nakedret: проверка naked returns в функциях длиннее 5 строк
	nakedRetConfig := &nakedret.NakedReturnRunner{
		MaxLength:     5,
		SkipTestFiles: false,
	}

	// Все стандартные анализаторы из golang.org/x/tools/go/analysis/passes
	checks := []*analysis.Analyzer{
		// Проверка корректности использования встроенной функции append
		appends.Analyzer,

		// Проверка корректности объявлений в assembly файлах
		asmdecl.Analyzer,

		// Обнаружение бесполезных присваиваний
		assign.Analyzer,

		// Проверка распространенных ошибок при использовании пакета sync/atomic
		atomic.Analyzer,

		// Проверка, что 64-битные атомарные операции выровнены
		atomicalign.Analyzer,

		// Обнаружение распространенных ошибок с булевыми операторами
		bools.Analyzer,

		// Построение SSA-представления (используется другими анализаторами)
		buildssa.Analyzer,

		// Проверка корректности build tags
		buildtag.Analyzer,

		// Обнаружение нарушений правил передачи указателей между Go и C
		cgocall.Analyzer,

		// Проверка незаполненных полей в композитных литералах
		composite.Analyzer,

		// Проверка случайного копирования блокировок (locks)
		copylock.Analyzer,

		// Построение графа потока управления (используется другими анализаторами)
		ctrlflow.Analyzer,

		// Проверка использования == и != с errors
		deepequalerrors.Analyzer,

		// Проверка распространенных ошибок в defer
		defers.Analyzer,

		// Проверка корректности директив компилятора
		directive.Analyzer,

		// Проверка, что второй аргумент errors.As является указателем на тип, реализующий error
		errorsas.Analyzer,

		// Обнаружение структур, которые могут быть оптимизированы путем перестановки полей
		fieldalignment.Analyzer,

		// Поиск вызовов определенных функций (используется для тестирования)
		findcall.Analyzer,

		// Проверка корректности использования frame pointer
		framepointer.Analyzer,

		// Проверка использования http.ServeMux
		httpmux.Analyzer,

		// Проверка, что тело http.Response закрывается
		httpresponse.Analyzer,

		// Обнаружение невозможных утверждений типа интерфейса
		ifaceassert.Analyzer,

		// Построение AST inspector (используется другими анализаторами)
		inspect.Analyzer,

		// Обнаружение ссылок на переменные цикла из замыканий
		loopclosure.Analyzer,

		// Проверка, что context.CancelFunc вызывается на всех путях выполнения
		lostcancel.Analyzer,

		// Проверка бесполезных сравнений с nil функций
		nilfunc.Analyzer,

		// Проверка на разыменование nil указателей
		nilness.Analyzer,

		// Проверка соответствия аргументов форматным строкам (Printf-like функции)
		printf.Analyzer,

		// Проверка случайного использования == и != с reflect.Value
		reflectvaluecompare.Analyzer,

		// Обнаружение затененных переменных
		shadow.Analyzer,

		// Проверка подозрительных сдвигов
		shift.Analyzer,

		// Обнаружение неправильного использования небуферизованных каналов с сигналами
		sigchanyzer.Analyzer,

		// Проверка неправильного использования slog
		slog.Analyzer,

		// Проверка вызовов sort.Slice с неправильной функцией сравнения
		sortslice.Analyzer,

		// Проверка сигнатур методов, известных стандартной библиотеке
		stdmethods.Analyzer,

		// Проверка, что используемые API доступны в целевой версии Go
		stdversion.Analyzer,

		// Проверка преобразований типа из int в string
		stringintconv.Analyzer,

		// Проверка корректности тегов структур
		structtag.Analyzer,

		// Проверка, что горутины в тестах не вызывают t.Fatal
		testinggoroutine.Analyzer,

		// Проверка распространенных ошибок в использовании пакета testing
		tests.Analyzer,

		// Проверка правильности использования time.Format
		timeformat.Analyzer,

		// Проверка передачи неверных типов в unmarshal функции
		unmarshal.Analyzer,

		// Обнаружение недостижимого кода
		unreachable.Analyzer,

		// Проверка неправильного использования unsafe.Pointer
		unsafeptr.Analyzer,

		// Проверка, что результаты вызовов определенных функций используются
		unusedresult.Analyzer,

		// Обнаружение записей в локальные переменные, которые никогда не читаются
		unusedwrite.Analyzer,

		// Обнаружение использования дженериков (используется другими анализаторами)
		usesgenerics.Analyzer,

		// Поиск naked returns в функциях длиннее 5 строк (nakedret)
		nakedret.NakedReturnAnalyzer(nakedRetConfig),

		// Проверка использования os.Exit в main функции main пакета (custom)
		OsExitCheckAnalyzer,
	}

	// Добавляем SA анализаторы из staticcheck
	for _, analyzer := range staticcheck.Analyzers {
		checks = append(checks, analyzer.Analyzer)
	}

	// Добавляем ST (Style) анализатор из stylecheck
	for _, analyzer := range stylecheck.Analyzers {
		checks = append(checks, analyzer.Analyzer)
		break
	}

	// Добавляем S (Simple) анализатор из simple
	for _, analyzer := range simple.Analyzers {
		checks = append(checks, analyzer.Analyzer)
		break
	}

	// Добавляем QF (Quickfix) анализатор из quickfix
	for _, analyzer := range quickfix.Analyzers {
		checks = append(checks, analyzer.Analyzer)
		break
	}

	multichecker.Main(checks...)
}
