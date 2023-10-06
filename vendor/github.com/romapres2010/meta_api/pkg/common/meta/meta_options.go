package meta

// QueryOptions опции запроса
type QueryOptions map[string]string

func (opt *QueryOptions) Copy() (out QueryOptions) {
	if opt != nil {
		out = make(QueryOptions, len(*opt))

		for key, val := range *opt {
			out[key] = val
		}
		return out
	}
	return nil
}

// Options опции расчета
type Options struct {
	Key                string               // Ключ, используются для поиска в cache, формат EntityName.EntityName.
	Entity             *Entity              // Обрабатываемая сущность
	FromEntity         *Entity              // Сущность из которой конвертировать
	CascadeUp          int                  // [cascade_up] сколько уровней вверх по FK
	CascadeDown        int                  // [cascade_down] сколько уровней вниз по FK
	DbOrder            string               // [db_order] последовательность сортировки строк в ответе
	DbWhere            string               // [db_where] фраза where для встраивания в запрос
	DbLimit            string               // [db_limit] ограничение на выборку данных в запросе
	DbOffset           string               // [db_offset] сдвиг строки, с которой начать выводить данные в запросе
	DbFieldsWhere      map[string]string    // Доп фильтрация по полям
	Fields             FieldsMap            // Ограничение на обрабатываемые поля
	FilterPreExpr      *Expr                // Правило фильтрации текущего slice в начале обработки (до вычисления всех полей)
	FilterPostExpr     *Expr                // Правило фильтрации текущего slice в конце обработки (после вычисления всех полей и выполнения всех внутренних фильтраций)
	FilterPostRefExprs map[*Reference]*Expr // Правила фильтрации вложенных Composition в конце обработки
	QueryOptions       QueryOptions         // опции запроса
	QueryOptionsDown   QueryOptions         // опции запроса для проброса ниже
	Global             *GlobalOptions       // глобальные опции расчета
}

func (opt *Options) Clone() (out *Options) {
	if opt != nil {
		out = &Options{}
		*out = *opt

		out.DbFieldsWhere = make(map[string]string, len(opt.DbFieldsWhere))
		for key, val := range opt.DbFieldsWhere {
			out.DbFieldsWhere[key] = val
		}

		out.FilterPostRefExprs = make(map[*Reference]*Expr, len(opt.FilterPostRefExprs))
		for key, val := range opt.FilterPostRefExprs {
			out.FilterPostRefExprs[key] = val
		}

		return out
	}
	return nil
}

// GlobalOptions глобальные опции расчета
type GlobalOptions struct {
	NameFormat             string // [name_format] формат именования полей в параметрах запроса 'json', 'yaml', 'xml', 'xsl', 'name'
	InFormat               string // [in_format] формат вывода результата 'json', 'yaml', 'xml', 'xsl'
	OutFormat              string // [out_format] формат вывода результата 'json', 'yaml', 'xml', 'xsl'
	TxExternal             uint64 // [tx] идентификатор внешней транзакции
	SkipCache              bool   // [skip_cache] принудительно считать из внешнего источника
	SkipCalculation        bool   // [skip_calculation] принудительно отключить все вычисления
	UseCache               bool   // [use_cache] принудительно использовать кеширование - имеет приоритет над skip_cache
	EmbedError             bool   // [embed_error] встраивать отдельные типы некритичных ошибок в текст ответа
	IgnoreExtraField       bool   // [ignore_extra_field] игнорировать лишние поля в параметрах запроса
	Validate               bool   // [validate] проверка данных
	OutTrace               bool   // [out_trace] вывод трассировки
	MultiRow               bool   // [multi_row] признак многострочной обработки
	StaticFiltering        bool   // [filter] признак статической фильтрации
	Persist                bool   // [persist] признак, что отправлять данные в хранилище
	KeepLock               bool   // признак блокировать данные в cache
	PersistRestrictFields  bool   // [persist_restrict_fields] ограничить поля сохранения, теми, что пришли на вход в JSON
	PersistUseUK           bool   // [persist_use_uk] для сохранения использовать UK, если не заполнен PK
	PersistUpdateAllFields bool   // [persist_update_all_fields] обновлять все поля объекта
}

func (opt *GlobalOptions) Clone() (out *GlobalOptions) {
	if opt != nil {
		out = &GlobalOptions{}
		*out = *opt
		return out
	}
	return nil
}
