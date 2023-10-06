package apiservice

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
)

type OptionsCache map[string]*_meta.Options

func contextWithOptionsCache(ctx context.Context) context.Context {
	optionCache := make(OptionsCache)
	return context.WithValue(ctx, ctxCacheKey, optionCache)
}

// fromContextOptionsCache extracts the OptionsCache from ctx, if present.
func fromContextOptionsCache(ctx context.Context) OptionsCache {
	if ctx != nil {
		optionsCache, ok := ctx.Value(ctxCacheKey).(OptionsCache)
		if !ok {
			return nil
		}
		return optionsCache
	}
	return nil
}

func fromContextTxId(ctx context.Context) uint64 {
	if ctx != nil {
		txId, ok := ctx.Value(ctxTxIdKey).(uint64)
		if !ok {
			return 0
		}
		return txId
	}
	return 0
}

func contextWithTxId(ctx context.Context, txId uint64) context.Context {
	return context.WithValue(ctx, ctxTxIdKey, txId)
}

func (s *Service) GetOptionFromCache(ctx context.Context, key string) *_meta.Options {
	if s != nil && ctx != nil {
		if optionsCache := fromContextOptionsCache(ctx); optionsCache != nil {
			if option, ok := optionsCache[key]; ok {
				return option
			}
		}
	}
	return nil
}

func (s *Service) SetOptionToCache(ctx context.Context, key string, options *_meta.Options) {
	if s != nil && ctx != nil {
		if optionCache := fromContextOptionsCache(ctx); optionCache != nil {
			optionCache[key] = options
		}
	}
}

func (s *Service) ParseQueryOptions(ctx context.Context, requestID uint64, key string, entity *_meta.Entity, referenceField *_meta.Field, queryOptions _meta.QueryOptions, gopt *_meta.GlobalOptions) (opt *_meta.Options, err error) {
	if s != nil && entity != nil {

		//_log.Debug("START: requestID, key, entityName", requestID, key, entity.Name)

		// Ищем в globalCache, переданном через context
		if opt = s.GetOptionFromCache(ctx, key); opt == nil {

			var (
				referenceName        string
				delimiterStart       = s.cfg.QueryOption.DelimiterStart
				delimiterEnd         = s.cfg.QueryOption.DelimiterEnd
				delimiterStartFilter = s.cfg.QueryOption.DelimiterStartFilter
			)

			opt = &_meta.Options{}
			opt.Key = key
			opt.Entity = entity
			opt.QueryOptions = queryOptions
			opt.Global = gopt // глобальные опции пришли на вход разобранными

			// Не чего формировать
			if queryOptions != nil {

				// Разберем глобальные опции один раз, если он пришел пустым на вход
				if gopt == nil {
					opt.Global = &_meta.GlobalOptions{}
					if err := s.parseGlobalOptions(requestID, entity, queryOptions, opt.Global); err != nil {
						return nil, err
					}
				}

				if fromEntityOption := queryOptions[s.cfg.QueryOption.FromEntityFull]; fromEntityOption != "" {
					if opt.FromEntity = s.GetEntityUnsafe(fromEntityOption); opt.Entity == nil {
						return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - '%s'='%s' does not exists", entity.Name, s.cfg.QueryOption.FromEntity, fromEntityOption)).PrintfError()
					}
				}

				// сколько уровней вверх по FK
				if cascadeUpOption := queryOptions[s.cfg.QueryOption.CascadeUpFull]; cascadeUpOption != "" {
					if opt.CascadeUp, err = strconv.Atoi(cascadeUpOption); err != nil {
						return nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Query parameter '%s' is incorrect integer '%s'", s.cfg.QueryOption.CascadeUp, cascadeUpOption))
					}
				}

				// сколько уровней вниз по FK
				if cascadeDownOption := queryOptions[s.cfg.QueryOption.CascadeDownFull]; cascadeDownOption != "" {
					if opt.CascadeDown, err = strconv.Atoi(cascadeDownOption); err != nil {
						return nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Query parameter '%s' is incorrect integer '%s'", s.cfg.QueryOption.CascadeDown, cascadeDownOption))
					}
				}

				// Имя ссылочного поля с "." на конце. Для корня - пустой referenceField - это будет ''
				if referenceField != nil {
					referenceName = referenceField.GetTagName(opt.Global.NameFormat, true) + "."
				} else {
					referenceName = ""
				}

				{ // [fields] Сформируем ограничение на поля
					var outFields []string
					var outFieldsDown []string // поля для проброса в reference

					// фильтрация список полей в ответе
					fieldsOption, ok := queryOptions[s.cfg.QueryOption.FieldsFull]
					if ok {
						// Если в запросе есть квалификатор [fields], но в нем пусто, то выводим только PK поля
						if fieldsOption != "" {
							fields := strings.Split(fieldsOption, ",")

							for _, field := range fields {
								fieldName := strings.TrimPrefix(field, referenceName)
								if strings.ContainsAny(fieldName, ".") {
									outFieldsDown = append(outFieldsDown, fieldName) // пробросим в reference
								} else {
									outFields = append(outFields, fieldName) // наверное, наше поле
								}
							}

							// Добавить поля того ключа, через который ссылаемся
							if referenceField != nil {
								if referenceField.Reference().ToKey() != nil {
									// Если поля не фильтруются, то выводить все и добавлять поля ключа не нужно
									if len(outFields) > 0 {
										for _, field := range referenceField.Reference().ToKey().Fields() {
											outFieldsDown = append(outFieldsDown, field.Name)
										}
									}
								} else {
									return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if referenceField.Reference().toKey != nil {}", []interface{}{referenceField.Reference()})
								}
							}

						} else {
							// Если пусто, то будут выбраны поля PK
							outFields = opt.Entity.PKFieldsName() // Только поля первичного ключа
						}
					} else {
						// Если нет квалификатора [fields], то все поля без ограничений
						outFields = nil
					}

					if len(outFields) > 0 {
						// Подготовим map с полями, которые нужны
						opt.Fields, err = s.constructFieldsMap(requestID, opt.Entity, outFields, opt.Global.NameFormat, opt.Global.IgnoreExtraField, false, false)
						if err != nil {
							return nil, err
						}
					}

					if len(outFieldsDown) > 0 {
						if opt.QueryOptionsDown == nil {
							opt.QueryOptionsDown = make(_meta.QueryOptions)
						}
						opt.QueryOptionsDown[s.cfg.QueryOption.FieldsFull] = strings.Join(outFieldsDown, ",") // Пробросить в reference
					}
				} // [fields] Сформируем ограничение на поля

				{ // Опции для БД

					opt.DbWhere = queryOptions[s.cfg.QueryOption.DbWhereFull]
					opt.DbLimit = queryOptions[s.cfg.QueryOption.DbLimitFull]
					opt.DbOffset = queryOptions[s.cfg.QueryOption.DbOffsetFull]

					{ // [db_order]
						var orderFields []string
						var orderFieldsDown []string

						orderOption := queryOptions[s.cfg.QueryOption.DbOrderFull] // правила сортировки
						if orderOption != "" {

							fields := strings.Split(orderOption, ",")

							for _, field := range fields {
								fieldName := strings.TrimPrefix(field, referenceName)
								if strings.ContainsAny(fieldName, ".") {
									orderFieldsDown = append(orderFieldsDown, fieldName) // пробросим в reference
								} else {
									orderFields = append(orderFields, fieldName) // наверное, наше поле
								}
							}

							if len(orderFields) > 0 {
								opt.DbOrder = strings.Join(orderFields, ",")
							}

							if len(orderFieldsDown) > 0 {
								if opt.QueryOptionsDown == nil {
									opt.QueryOptionsDown = make(_meta.QueryOptions)
								}
								opt.QueryOptionsDown[s.cfg.QueryOption.DbOrderFull] = strings.Join(orderFieldsDown, ",") // Пробросить в reference
							}
						}
					} // [db_order]

				} // Опции для БД

				{ // Условия поиска по полям
					for optName, optVal := range queryOptions {
						if len(optVal) != 0 {

							// все что не в [] трактуем как поля для дополнительной фильтрации запроса
							if !(strings.ContainsAny(optName, delimiterStart) || strings.ContainsAny(optName, delimiterEnd)) {

								fieldName := strings.TrimPrefix(optName, referenceName)
								if strings.ContainsAny(fieldName, ".") {
									if opt.QueryOptionsDown == nil {
										opt.QueryOptionsDown = make(_meta.QueryOptions)
									}
									opt.QueryOptionsDown[fieldName] = optVal
								} else {
									// Наверное, наше поле - далее проверим
									if opt.DbFieldsWhere == nil {
										opt.DbFieldsWhere = make(map[string]string)
									}
									opt.DbFieldsWhere[fieldName] = optVal
								}
							}
						}
					}
				} // Условия поиска по полям

				{ // Сформируем правила фильтрации Composition
					for optName, optVal := range queryOptions {
						if len(optVal) != 0 {

							// Если поле начинается с '[filter', то трактуем как доп параметры фильтрации Composition
							if strings.HasPrefix(optName, delimiterStartFilter) {

								// Имя композиции, может быть иерархическим - сформируем смысловую часть
								compositionName := strings.TrimSuffix(strings.TrimPrefix(optName, delimiterStartFilter), delimiterEnd)

								if compositionName != "" {

									// Через "." после [filter указываются правила фильтрации на Composition
									compositionName = strings.TrimPrefix(compositionName, ".")
									compositionName = strings.TrimPrefix(compositionName, referenceName)

									if compositionName == "" {

										// Если после "." ни чего нет, то это означает фильтрацию текущего объекта - POST
										filterCodeFull := "filter(" + opt.Entity.Name + ", {" + optVal + "})"

										// Создадим и инициируем выражение фильтрации
										opt.FilterPreExpr, err = _meta.NewExpr(opt.Entity, nil, _meta.STATUS_ENABLED, "Filter inside get Entity", _meta.EXPR_FILTER, _meta.EXPR_ACTION_INSIDE_GET, filterCodeFull, nil, nil, true)
										if err != nil {
											return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - error prepare filter rule filterCodeFull='%s'", opt.Entity.Name, filterCodeFull)).PrintfError()
										}

									} else if !strings.ContainsAny(compositionName, ".") {

										// Если точки нет - то проверяем Composition текущей сущности, иначе это относится к подчиненным объектам
										if composition := opt.Entity.GetComposition(compositionName); composition != nil {

											filterCodeFull := "filter(" + composition.Field().Name + ", {" + optVal + "})"

											// Создадим и инициируем выражение фильтрации
											expr, err := _meta.NewExpr(composition.Entity(), composition.Field(), _meta.STATUS_ENABLED, "Filter post get Composition '"+compositionName+"'", _meta.EXPR_FILTER, _meta.EXPR_ACTION_POST_GET, filterCodeFull, []string{composition.Field().Name}, nil, true)
											if err != nil {
												return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - error prepare filter rule for Composition '%s', filterCode='%s'", opt.Entity.Name, compositionName, filterCodeFull)).PrintfError()
											}

											if opt.FilterPostRefExprs == nil {
												opt.FilterPostRefExprs = make(map[*_meta.Reference]*_meta.Expr)
											}

											opt.FilterPostRefExprs[composition] = expr

										} else {
											// Не существует такой Composition
											return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - does not have Composition with name '%s', urlOption='%s=%s'", entity.Name, compositionName, optName, optVal)).PrintfError()
										}
									} else {

										// относится к подчиненным объектам - пробросим через рекурсию на уровень ниже
										optFilterName := delimiterStartFilter + "." + compositionName + delimiterEnd // Сформируем правила для фильтрации
										if opt.QueryOptionsDown == nil {
											opt.QueryOptionsDown = make(_meta.QueryOptions)
										}
										opt.QueryOptionsDown[optFilterName] = optVal
									}
								} else {

									// после [filter] без точек указана Post фильтрация всей сущности
									filterCodeFull := "filter(" + opt.Entity.Name + ", {" + optVal + "})"

									// Создадим и инициируем выражение фильтрации
									opt.FilterPostExpr, err = _meta.NewExpr(opt.Entity, nil, _meta.STATUS_ENABLED, "Filter post get Entity", _meta.EXPR_FILTER, _meta.EXPR_ACTION_POST_GET, filterCodeFull, nil, nil, true)
									if err != nil {
										return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - error prepare filter rule filterCodeFull='%s'", opt.Entity.Name, filterCodeFull)).PrintfError()
									}
								}
							}
						}
					}
				} // Сформируем правила фильтрации Composition
			}

			// Поместим в globalCache
			s.SetOptionToCache(ctx, key, opt)
		}

		return opt, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil {}", []interface{}{s, entity}).PrintfError()
}

func (s *Service) parseGlobalOptions(requestID uint64, entity *_meta.Entity, queryOptions _meta.QueryOptions, gopt *_meta.GlobalOptions) (err error) {
	if s != nil && queryOptions != nil && gopt != nil && entity != nil {

		var (
			skipCacheOption              = queryOptions[s.cfg.QueryOption.SkipCacheFull]              // принудительно считать из внешнего источника
			useCacheOption               = queryOptions[s.cfg.QueryOption.UseCacheFull]               // принудительно использовать кеширование - имеет приоритет над skip_cache
			embedErrorOption             = queryOptions[s.cfg.QueryOption.EmbedErrorFull]             // встраивать отдельные типы некритичных ошибок в текст ответа
			validateOption               = queryOptions[s.cfg.QueryOption.ValidateFull]               // выполнять валидация данных
			txExternalOption             = queryOptions[s.cfg.QueryOption.TxExternalFull]             // идентификатор внешней транзакции
			ignoreExtraField             = queryOptions[s.cfg.QueryOption.IgnoreExtraFieldFull]       // игнорировать лишние поля в параметрах запроса
			nameFormatOption             = queryOptions[s.cfg.QueryOption.NameFormatFull]             // формат именования полей в параметрах запроса 'json', 'yaml', 'xml', 'xsl', 'name'
			outFormatOption              = queryOptions[s.cfg.QueryOption.OutFormatFull]              // формат вывода результата 'json', 'yaml', 'xml', 'xsl'
			outTraceOption               = queryOptions[s.cfg.QueryOption.OutTraceFull]               // вывод трассировки
			multiRowOption               = queryOptions[s.cfg.QueryOption.MultiRowFull]               // признак многострочной обработки
			staticFilteringOption        = queryOptions[s.cfg.QueryOption.StaticFilteringFull]        // признак статической фильтрации
			persistOption                = queryOptions[s.cfg.QueryOption.PersistFull]                // признак, что отправлять данные в хранилище
			skipCalculationOption        = queryOptions[s.cfg.QueryOption.SkipCalculationFull]        // принудительно отключить все вычисления
			persistRestrictFieldsOption  = queryOptions[s.cfg.QueryOption.PersistRestrictFieldsFull]  // ограничить поля сохранения, теми, что пришли на вход в Marshal
			persistUseUKOption           = queryOptions[s.cfg.QueryOption.PersistUseUKFull]           // для сохранения использовать UK, если не заполнен PK
			persistUpdateAllFieldsOption = queryOptions[s.cfg.QueryOption.PersistUpdateAllFieldsFull] // обновлять все поля объекта
		)

		_log.Debug("START: requestID", requestID)

		{ // формат в котором заданы поля
			// Если не задано, то формат берем из запроса, иначе "name"
			if nameFormatOption == "" {
				nameFormatOption = "name" // по умолчанию в именах полей
			}

			switch nameFormatOption {
			case "json", "xml", "yaml", "xls", "name":
				gopt.NameFormat = nameFormatOption
			default:
				return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, fmt.Sprintf("Allowed only query parameter '%s'='json', 'yaml', 'xml', 'xsl', 'name'", s.cfg.QueryOption.NameFormat)).PrintfError()
			}
		} // формат в котором заданы поля

		{ // формат вывода результата
			switch outFormatOption {
			case "json", "xml", "yaml", "xls":
				gopt.OutFormat = outFormatOption
			default:
				return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, fmt.Sprintf("Allowed only query parameter '%s'='json', 'yaml', 'xml', 'xsl'", s.cfg.QueryOption.OutFormat)).PrintfError()
			}
		} // формат вывода результата

		// Кеширование может быть запрещено на уровне отдельного запроса или для сущности в целом, принудительно включение кеширование для отдельного запроса имеет приоритет
		gopt.UseCache = !(skipCacheOption == "true" || entity.SkipCache) || useCacheOption == "true"

		gopt.IgnoreExtraField = ignoreExtraField == "true"

		gopt.EmbedError = embedErrorOption == "true"

		gopt.MultiRow = multiRowOption == "true"

		gopt.StaticFiltering = staticFilteringOption == "true"

		gopt.Persist = persistOption == "true"

		gopt.PersistRestrictFields = persistRestrictFieldsOption == "true"

		gopt.PersistUseUK = persistUseUKOption == "true"

		gopt.PersistUpdateAllFields = persistUpdateAllFieldsOption == "true"

		gopt.SkipCalculation = skipCalculationOption == "true"

		gopt.OutTrace = outTraceOption == "true"

		gopt.Validate = validateOption == "true" // валидация по умолчанию выключена

		// идентификатор внешней транзакции
		if txExternalOption != "" {
			if gopt.TxExternal, err = strconv.ParseUint(txExternalOption, 10, 64); err != nil {
				return _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Query parameter '%s' is incorrect uinteger '%s'", s.cfg.QueryOption.TxExternal, txExternalOption))
			}
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && queryOptions != nil && gopt != nil && entity != nil {}", []interface{}{s, queryOptions, gopt, entity}).PrintfError()
}

// constructFieldsMap сформировать набор полей для ограничения возвращаемых полей
func (s *Service) constructFieldsMap(requestID uint64, entity *_meta.Entity, fieldsName []string, nameFormat string, ignoreExtra bool, addRefFields bool, addUKFields bool) (fieldsMap _meta.FieldsMap, err error) {
	if s != nil && entity != nil {
		isRestrictFieldsTag := fieldsName != nil && len(fieldsName) > 0

		if isRestrictFieldsTag {
			var errDetail []string // ошибки агрегируем, чтобы потом отдать все сразу

			for _, fieldName := range fieldsName {

				// Ищем поля с учетом формата имени поля
				if field := entity.FieldByTagName(nameFormat, fieldName); field != nil {

					// создадим в момент первого использования, так как может и не существовать запрошенного поля
					if fieldsMap == nil {
						fieldsMap = make(_meta.FieldsMap, len(entity.StructFields())) // максимальное количество равно всем полям структуры
					}
					fieldsMap[field.Name] = field

					// Если поле является ссылкой, то нужно добавить все поля этой ссылки, чтобы они выбирались из БД
					if field.InternalType == _meta.FIELD_TYPE_ASSOCIATION || field.InternalType == _meta.FIELD_TYPE_COMPOSITION {

						ref := field.Reference()
						if ref != nil {
							ref.AddFieldsToMap(&fieldsMap)
						} else {
							errDetail = append(errDetail, fmt.Sprintf("Entity '%s', Reference '%s' - empty 'Reference' pointer", entity.Name, field.Name))
						}
					}

					// Если поле входит в состав ссылки, то нужно добавить эту ссылку как поле и все поля этой ссылки
					if addRefFields {
						for _, ref := range field.References() {
							if ref != nil {
								ref.AddFieldsToMap(&fieldsMap)
							} else {
								errDetail = append(errDetail, fmt.Sprintf("Entity '%s', Reference '%s' - empty 'Reference' pointer", entity.Name, field.Name))
							}
						}
					}

				} else {
					// не найденные поля игнорируем, если в них есть "." - это может относиться к вложенным сущностям
					if !(ignoreExtra || strings.ContainsAny(fieldName, ".")) {
						errDetail = append(errDetail, fmt.Sprintf("Entity '%s' has not have enabled field '%s', '%s'='%s'", entity.Name, fieldName, s.cfg.QueryOption.NameFormat, nameFormat))
					}
				}
			}

			// Добавим реверсивную ссылку Composition-Association
			if addRefFields {
				// создадим в момент первого использования, так как может и не существовать запрошенного поля
				if fieldsMap == nil {
					fieldsMap = make(_meta.FieldsMap, len(entity.StructFields())) // максимальное количество равно всем полям структуры
				}

				for _, ref := range entity.References() {
					if ref.Type == _meta.REFERENCE_TYPE_ASSOCIATION {
						if refField := ref.Field(); refField != nil {
							fieldsMap[refField.Name] = refField
						}
					}
				}
			}

			if addUKFields {
				// создадим в момент первого использования, так как может и не существовать запрошенного поля
				if fieldsMap == nil {
					fieldsMap = make(_meta.FieldsMap, len(entity.StructFields())) // максимальное количество равно всем полям структуры
				}
				for _, key := range entity.KeysUK() {
					//if key.Type == _meta.KEY_TYPE_UK {
					key.AddFieldsToMap(&fieldsMap)
					//}
				}
			}

			// Обработаем накопленные ошибки
			if len(errDetail) > 0 {
				errMy := _err.NewTyped(_err.ERR_ERROR, requestID, "Error process Entity fields")
				errMy.Details = errDetail
				// вернем ожидаемый формат структуры
				if rowExpected, err := s.NewRowAll(requestID, entity, nil); err == nil {
					valJson, _ := json.MarshalIndent(rowExpected.Value, "", "    ")
					outMes := fmt.Sprintf("{\"expected_fields\": " + string(valJson) + "}")
					errMy.MessageJson = []byte(outMes)
					//errMy.Extra = rowExpected.Value
				}
				return nil, errMy
			}

		} else {
			// Если нет ограничения на состав полей, то передаем все
			//fieldsMap = entity.StructFieldsMap() // TODO - копия снижает производительность. Проверить затраты
			//deepcopy.Clone()
			return nil, nil
		}

		return fieldsMap, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil {}", []interface{}{s, entity}).PrintfError()
}
