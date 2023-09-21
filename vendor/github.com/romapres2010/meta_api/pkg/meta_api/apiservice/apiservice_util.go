package apiservice

//
//type QueryOptionsParse struct {
//    QueryOptions     _meta.QueryOptions // опции переданные через запрос
//    Entity           *_meta.Entity      // Обрабатываемая сущность
//    FromEntity       *_meta.Entity      // Сущность из которой конвертировать
//    ArgsFields           _meta.FieldsMap    // Ограничение на обрабатываемые поля
//    FromEntityName   string             // Имя входной сущности
//    UseCache         bool               // принудительно считать из внешнего источника
//    EmbedError       bool               // встраивать отдельные типы некритичных ошибок в текст ответа
//    Validate         bool               // проверка данных
//    IgnoreExtraField bool               // игнорировать лишние поля в json
//    CascadeUp        int                // сколько уровней вверх
//    CascadeDown      int                // сколько уровней вниз
//    TxExternal       uint64             // идентификатор внешней транзакции
//    NameFormat       string             // формат именования полей в URL запроса
//    OutFormat        string             // формат вывода результата 'json', 'yaml', 'xml', 'xsl'
//    OutTrace         bool               // вывод трассировки
//    MultiRow         bool               // признак многострочной обработки
//    StaticFiltering  bool               // признак статической фильтрации
//    FilterCode       string             // правила фильтрации текущего объекта
//    FilterCodeMap    map[string]string  // правила фильтрации вложенных Composition
//}
//
//func (qopt *QueryOptionsParse) prepareProcessOptions() (popt *_meta.Options, err error) {
//    if qopt != nil {
//        popt = &_meta.Options{}
//
//        //popt.Entity = qopt.Entity
//        popt.TxExternal = qopt.TxExternal
//        popt.UseCache = qopt.UseCache
//        popt.EmbedError = qopt.EmbedError
//        popt.Validate = qopt.Validate
//        popt.OutTrace = qopt.OutTrace
//        popt.MultiRow = qopt.MultiRow
//        popt.ArgsFields = qopt.ArgsFields
//        popt.StaticFiltering = qopt.StaticFiltering
//
//        { // Обработаем дополнительные условия фильтрации текущей сущности
//            if qopt.FilterCode != "" {
//                filterCodeFull := "filter(" + qopt.Entity.Name + ", {" + qopt.FilterCode + "})"
//
//                // Создадим и инициируем выражение фильтрации
//                popt.FilterPreExpr, err = _meta.NewExpr(qopt.Entity, nil, _meta.STATUS_ENABLED, "Filter inside get", _meta.EXPR_FILTER, _meta.EXPR_ACTION_INSIDE_GET, filterCodeFull, nil, true)
//                if err != nil {
//                    return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - error prepare filter rule filterCodeFull='%s'", qopt.Entity.Name, filterCodeFull)).PrintfError()
//                }
//            }
//        } // Обработаем дополнительные условия фильтрации текущей сущности
//
//        { // Обработаем дополнительные условия фильтрации Composition
//            for compositionName, filterCode := range qopt.FilterCodeMap {
//                if composition := qopt.Entity.GetComposition(compositionName); composition != nil {
//
//                    filterCodeFull := "filter(" + composition.Field().Name + ", {" + filterCode + "})"
//
//                    // Создадим и инициируем выражение фильтрации
//                    expr, err := _meta.NewExpr(composition.Entity(), composition.Field(), _meta.STATUS_ENABLED, "Filter post get Composition '"+compositionName+"'", _meta.EXPR_FILTER, _meta.EXPR_ACTION_POST_GET, filterCodeFull, []string{composition.Field().Name}, true)
//                    if err != nil {
//                        return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - error prepare filter rule for Composition '%s', filterCode='%s'", qopt.Entity.Name, compositionName, filterCodeFull)).PrintfError()
//                    }
//
//                    if popt.FilterPostRefExprs == nil {
//                        popt.FilterPostRefExprs = make(map[*_meta.Reference]*_meta.Expr)
//                    }
//
//                    popt.FilterPostRefExprs[composition] = expr
//                }
//            }
//        } // Обработаем дополнительные условия фильтрации Composition
//
//        return popt, nil
//    }
//    return nil, nil
//}
//
//func (s *Service) parseQueryOptions(ctx context.Context, requestID uint64, entityName string, inFormat string, queryOptions _meta.QueryOptions) (qopt *QueryOptionsParse, popt *_meta.Options, err error) {
//    if s != nil && s.storageMap != nil && entityName != "" && queryOptions != nil {
//
//        var (
//            //delimiterStart       = s.cfg.QueryOption.DelimiterStart
//            delimiterEnd         = s.cfg.QueryOption.DelimiterEnd
//            delimiterStartFilter = s.cfg.QueryOption.DelimiterStartFilter
//            cascadeUpOption      = queryOptions[s.cfg.QueryOption.CascadeUpFull]   // сколько уровней вверх по FK
//            cascadeDownOption    = queryOptions[s.cfg.QueryOption.CascadeDownFull] // сколько уровней вниз по FK
//            fromEntityOption     = queryOptions[s.cfg.QueryOption.FromEntityFull]  // имя входной сущности
//        )
//
//        qopt = &QueryOptionsParse{}
//        qopt.QueryOptions = queryOptions
//
//        _log.Debug("START: requestID, entityName", requestID, entityName)
//
//        if qopt.Entity = s.getEntityUnsafe(entityName); qopt.Entity == nil {
//            return nil, nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity with name '%s' does not exists", entityName)).PrintfError()
//        }
//
//        if fromEntityOption != "" {
//
//            qopt.FromEntityName = fromEntityOption
//
//            if qopt.FromEntity = s.getEntityUnsafe(fromEntityOption); qopt.Entity == nil {
//                return nil, nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - '%s'='%s' does not exists", entityName, s.cfg.QueryOption.FromEntity, fromEntityOption)).PrintfError()
//            }
//        }
//
//        // Разберем базовые опции
//        if err := s.parseQueryOptionsGlobal(requestID, queryOptions, qopt); err != nil {
//            return nil, nil, err
//        }
//
//        // сколько уровней вверх по FK
//        if cascadeUpOption != "" {
//            if qopt.CascadeUp, err = strconv.Atoi(cascadeUpOption); err != nil {
//                return nil, nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Query parameter '%s' is incorrect integer '%s'", s.cfg.QueryOption.CascadeUp, cascadeUpOption))
//            }
//        }
//
//        // сколько уровней вниз по FK
//        if cascadeDownOption != "" {
//            if qopt.CascadeDown, err = strconv.Atoi(cascadeDownOption); err != nil {
//                return nil, nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Query parameter '%s' is incorrect integer '%s'", s.cfg.QueryOption.CascadeDown, cascadeDownOption))
//            }
//        }
//
//        { // Сформируем ограничение на поля
//            var outFields []string
//            // фильтрация список полей в ответе
//            fieldsOption, ok := queryOptions[s.cfg.QueryOption.FieldsFull]
//            if ok {
//                // Если в запросе есть квалификатор [fields], но в нем пусто, то выводим только PK поля
//                if fieldsOption != "" {
//                    outFields = strings.Split(fieldsOption, ",")
//                } else {
//                    // Если пусто, то будут выбраны поля PK
//                    outFields = qopt.Entity.PKFieldsName() // Только поля первичного ключа
//                }
//            } else {
//                // Если нет квалификатора [fields], то все поля без ограничений
//                outFields = nil
//            }
//
//            // Если нет квалификатора [fields], то все поля без ограничений
//            if qopt.ArgsFields, err = s.constructFieldsMap(requestID, qopt.Entity, outFields, qopt.NameFormat, qopt.IgnoreExtraField, false); err != nil {
//                return nil, nil, err
//            }
//
//        } // Сформируем ограничение на поля
//
//        { // Сформируем правила фильтрации Composition
//            for optName, optVal := range queryOptions {
//                if len(optVal) != 0 {
//                    // Если поле начинается с [filter, то трактуем как доп параметры фильтрации Composition
//                    if strings.HasPrefix(optName, delimiterStartFilter) {
//
//                        // Имя композиции, может быть иерархическим - сформируем смысловую часть
//                        compositionName := strings.TrimSuffix(strings.TrimPrefix(optName, delimiterStartFilter), delimiterEnd)
//
//                        if compositionName != "" {
//                            // Через "." указываются правила фильтрации на Composition
//                            compositionName = strings.TrimPrefix(compositionName, ".")
//
//                            if compositionName == "" {
//                                // Если после "." ни чего нет, то это означает фильтрацию текущего объекта
//                                qopt.FilterCode = optVal
//
//                            } else if !strings.ContainsAny(compositionName, ".") {
//                                // Если точки нет - то проверяем Composition текущей сущности, иначе это относится к подчиненным объектам
//                                // TODO - выделить общую часть из parseQueryOptionsReference
//                                if qopt.Entity.GetComposition(compositionName) != nil {
//                                    // Сохраним код фильтрации Composition
//                                    if qopt.FilterCodeMap == nil {
//                                        qopt.FilterCodeMap = make(map[string]string)
//                                    }
//                                    qopt.FilterCodeMap[compositionName] = optVal
//                                } else {
//                                    // Не существует такой Composition
//                                    return nil, nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - does not have Composition with name '%s', urlOption='%s=%s'", entityName, compositionName, optName, optVal)).PrintfError()
//                                }
//                                // TODO - выделить общую часть из parseQueryOptionsReference
//                            }
//                        }
//                    }
//                }
//            }
//        } // Сформируем правила фильтрации Composition
//
//        if popt, err = qopt.prepareProcessOptions(); err != nil {
//            return nil, nil, err
//        }
//
//        return qopt, popt, nil
//    }
//    return nil, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.storageMap != nil && entityName != '' && queryOptions != nil {}", []interface{}{s, entityName, queryOptions}).PrintfError()
//}
//
//func (s *Service) parseQueryOptionsGlobal(requestID uint64, queryOptions _meta.QueryOptions, qopt *QueryOptionsParse) (err error) {
//    if s != nil && s.storageMap != nil && queryOptions != nil {
//
//        var (
//            skipCacheOption       = queryOptions[s.cfg.QueryOption.SkipCacheFull]        // принудительно считать из внешнего источника
//            useCacheOption        = queryOptions[s.cfg.QueryOption.UseCacheFull]         // принудительно использовать кеширование - имеет приоритет над skip_cache
//            embedErrorOption      = queryOptions[s.cfg.QueryOption.EmbedErrorFull]       // встраивать отдельные типы некритичных ошибок в текст ответа
//            validateOption        = queryOptions[s.cfg.QueryOption.ValidateFull]         // выполнять валидация данных
//            txExternalOption      = queryOptions[s.cfg.QueryOption.TxExternalFull]       // идентификатор внешней транзакции
//            ignoreExtraField      = queryOptions[s.cfg.QueryOption.IgnoreExtraFieldFull] // игнорировать лишние поля в параметрах запроса
//            nameFormatOption      = queryOptions[s.cfg.QueryOption.NameFormatFull]       // формат именования полей в параметрах запроса 'json', 'yaml', 'xml', 'xsl', 'name'
//            outFormatOption       = queryOptions[s.cfg.QueryOption.OutFormatFull]        // формат вывода результата 'json', 'yaml', 'xml', 'xsl'
//            outTraceOption        = queryOptions[s.cfg.QueryOption.OutTraceFull]         // вывод трассировки
//            multiRowOption        = queryOptions[s.cfg.QueryOption.MultiRowFull]         // признак многострочной обработки
//            staticFilteringOption = queryOptions[s.cfg.QueryOption.StaticFilteringFull]  // признак статической фильтрации
//        )
//
//        _log.Debug("START: requestID", requestID)
//
//        { // формат в котором заданы поля
//            // Если не задано, то формат берем из запроса, иначе "name"
//            if nameFormatOption == "" {
//                nameFormatOption = "name" // по умолчанию в именах полей
//            }
//
//            switch nameFormatOption {
//            case "json", "xml", "yaml", "xls", "name":
//                qopt.NameFormat = nameFormatOption
//            default:
//                return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, fmt.Sprintf("Allowed only query parameter '%s'='json', 'yaml', 'xml', 'xsl', 'name'", s.cfg.QueryOption.NameFormat)).PrintfError()
//            }
//        } // формат в котором заданы поля
//
//        { // формат вывода результата
//            switch outFormatOption {
//            case "json", "xml", "yaml", "xls":
//                qopt.OutFormat = outFormatOption
//            default:
//                return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, fmt.Sprintf("Allowed only query parameter '%s'='json', 'yaml', 'xml', 'xsl'", s.cfg.QueryOption.OutFormat)).PrintfError()
//            }
//        } // формат вывода результата
//
//        // Кеширование может быть запрещено на уровне отдельного запроса или для сущности в целом, принудительно включение кеширование для отдельного запроса имеет приоритет
//        qopt.UseCache = !(skipCacheOption == "true" || qopt.Entity.SkipCache) || useCacheOption == "true"
//
//        qopt.IgnoreExtraField = ignoreExtraField == "true"
//
//        qopt.EmbedError = embedErrorOption == "true"
//
//        qopt.MultiRow = multiRowOption == "true"
//
//        qopt.StaticFiltering = staticFilteringOption == "true"
//
//        qopt.OutTrace = outTraceOption == "true"
//
//        qopt.Validate = validateOption == "true" // валидация по умолчанию выключена
//
//        // идентификатор внешней транзакции
//        if txExternalOption != "" {
//            if qopt.TxExternal, err = strconv.ParseUint(txExternalOption, 10, 64); err != nil {
//                return _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Query parameter '%s' is incorrect uinteger '%s'", s.cfg.QueryOption.TxExternal, txExternalOption))
//            }
//        }
//
//        return nil
//    }
//    return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.storageMap != nil && queryOptions != nil {}", []interface{}{s, queryOptions}).PrintfError()
//}
//
//// parseQueryOptionsReference - сформировать опции запроса для каскадного вызова
//func (s *Service) parseQueryOptionsReference(ctx context.Context, requestID uint64, referenceEntity *_meta.Entity, referenceField *_meta.Field, queryOptionsIn _meta.QueryOptions) (qopt *QueryOptionsParse, popt *_meta.Options, err error) {
//    if s != nil && s.storageMap != nil && referenceEntity != nil && queryOptionsIn != nil && referenceField != nil && referenceField.Reference() != nil {
//
//        var (
//            queryOptionsOut      = make(map[string]string)
//            delimiterStart       = s.cfg.QueryOption.DelimiterStart
//            delimiterEnd         = s.cfg.QueryOption.DelimiterEnd
//            delimiterStartFilter = s.cfg.QueryOption.DelimiterStartFilter
//            fieldsOption         = queryOptionsIn[s.cfg.QueryOption.FieldsFull]  // список полей
//            orderOption          = queryOptionsIn[s.cfg.QueryOption.DbOrderFull] // правила сортировки
//        )
//
//        qopt = &QueryOptionsParse{}
//        qopt.Entity = referenceEntity
//        qopt.QueryOptions = queryOptionsOut
//
//        _log.Debug("START: requestID, entityName", requestID, referenceEntity.Name)
//
//        // Разберем базовые опции
//        // TODO - перепроектировать передачу опций
//        if err := s.parseQueryOptionsGlobal(requestID, queryOptionsIn, qopt); err != nil {
//            return nil, nil, err
//        }
//
//        { // Часть глобальных параметров пробросить в queryOptionsOut
//            // TODO - перепроектировать
//            if val, ok := queryOptionsIn[s.cfg.QueryOption.UseCacheFull]; ok {
//                queryOptionsOut[s.cfg.QueryOption.UseCacheFull] = val
//            }
//            if val, ok := queryOptionsIn[s.cfg.QueryOption.EmbedErrorFull]; ok {
//                queryOptionsOut[s.cfg.QueryOption.EmbedErrorFull] = val
//            }
//            if val, ok := queryOptionsIn[s.cfg.QueryOption.ValidateFull]; ok {
//                queryOptionsOut[s.cfg.QueryOption.ValidateFull] = val
//            }
//            if val, ok := queryOptionsIn[s.cfg.QueryOption.IgnoreExtraFieldFull]; ok {
//                queryOptionsOut[s.cfg.QueryOption.IgnoreExtraFieldFull] = val
//            }
//            if val, ok := queryOptionsIn[s.cfg.QueryOption.TxExternalFull]; ok {
//                queryOptionsOut[s.cfg.QueryOption.TxExternalFull] = val
//            }
//            if val, ok := queryOptionsIn[s.cfg.QueryOption.NameFormatFull]; ok {
//                queryOptionsOut[s.cfg.QueryOption.NameFormatFull] = val
//            }
//            if val, ok := queryOptionsIn[s.cfg.QueryOption.OutFormatFull]; ok {
//                queryOptionsOut[s.cfg.QueryOption.OutFormatFull] = val
//            }
//            if val, ok := queryOptionsIn[s.cfg.QueryOption.OutTraceFull]; ok {
//                queryOptionsOut[s.cfg.QueryOption.OutTraceFull] = val
//            }
//            if val, ok := queryOptionsIn[s.cfg.QueryOption.StaticFilteringFull]; ok {
//                queryOptionsOut[s.cfg.QueryOption.StaticFilteringFull] = val
//            }
//        } // Часть глобальных параметров пробросить в queryOptionsOut
//
//        if referenceField.Reference().ToEntity() != referenceEntity {
//            return nil, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if referenceField.Reference().toEntity != referenceEntity {}", []interface{}{referenceField.Reference().ToEntity(), referenceEntity})
//        }
//
//        // Имя ссылочного поля с "." на конце
//        referenceFieldName := referenceField.GetTagName(qopt.NameFormat, true) + "."
//
//        { // [db_order] для вложенной сущности
//            if orderOption != "" {
//                orderFields := strings.Split(orderOption, ",")
//                var orderRefFields []string
//                for _, orderField := range orderFields {
//                    // Если наша сущность присутствует, то скопируем сортировку для под запроса
//                    if strings.HasPrefix(orderField, referenceFieldName) {
//                        fieldName := strings.TrimPrefix(orderField, referenceFieldName)
//                        orderRefFields = append(orderRefFields, fieldName)
//                    }
//                }
//                if len(orderRefFields) > 0 {
//                    queryOptionsOut[s.cfg.QueryOption.DbOrderFull] = strings.Join(orderRefFields, ",")
//                } // параметры сортировки
//            }
//        } // [db_order] для вложенной сущности
//
//        { // [fields] для вложенной сущности
//            if fieldsOption != "" {
//                var outFields []string
//                fields := strings.Split(fieldsOption, ",")
//
//                // Если наша сущность присутствует, то скопируем поля для под запроса
//                for _, field := range fields {
//                    if strings.HasPrefix(field, referenceFieldName) {
//                        fieldName := strings.TrimPrefix(field, referenceFieldName)
//                        outFields = append(outFields, fieldName)
//                    }
//                }
//
//                // Добавит поля того ключа, через который ссылаемся
//                if referenceField.Reference().ToKey() != nil {
//                    // Если поля не фильтруются, то выводить все и добавлять поля ключа не нужно
//                    if len(outFields) > 0 {
//                        for _, field := range referenceField.Reference().ToKey().ArgsFields() {
//                            outFields = append(outFields, field.Name)
//                        }
//                    }
//                } else {
//                    return nil, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if referenceField.Reference().toKey != nil {}", []interface{}{referenceField.Reference()})
//                }
//
//                if len(outFields) > 0 {
//                    queryOptionsOut[s.cfg.QueryOption.FieldsFull] = strings.Join(outFields, ",")
//                    // Подготовим map с полями, которые нужны
//                    qopt.ArgsFields, err = s.constructFieldsMap(requestID, qopt.Entity, outFields, qopt.NameFormat, qopt.IgnoreExtraField, false)
//                    if err != nil {
//                        return nil, nil, err
//                    }
//                }
//            }
//        } // [fields] для вложенной сущности
//
//        { // Добавим условия поиска по остальным полям для вложенной сущности
//            for optName, optVal := range queryOptionsIn {
//                if len(optVal) != 0 {
//                    // все что не в [] трактуем как поля для дополнительной фильтрации запроса
//                    if !(strings.ContainsAny(optName, delimiterStart) || strings.ContainsAny(optName, delimiterEnd)) {
//
//                        // Если имя ссылочного поля присутствует, то скопируем поля для под запроса
//                        if strings.HasPrefix(optName, referenceFieldName) {
//                            fieldName := strings.TrimPrefix(optName, referenceFieldName)
//                            queryOptionsOut[fieldName] = optVal
//                        }
//                    }
//                }
//            }
//        } // Добавим условия поиска по остальным полям для вложенной сущности
//
//        { // Сформируем правила фильтрации Composition
//            for optName, optVal := range queryOptionsIn {
//                if len(optVal) != 0 {
//                    // Если поле начинается с [filter, то трактуем как доп параметры фильтрации Composition
//                    if strings.HasPrefix(optName, delimiterStartFilter) {
//
//                        // Имя композиции, может быть иерархическим - сформируем смысловую часть
//                        compositionName := strings.TrimSuffix(strings.TrimPrefix(optName, delimiterStartFilter), delimiterEnd)
//
//                        if compositionName != "" {
//                            // Через "." после [filter указываются правила фильтрации на Composition
//                            compositionName = strings.TrimPrefix(compositionName, ".")
//
//                            // Если точка есть - то проверяем Composition текущей сущности и формируем правила фильтрации для подчиненных
//                            if strings.ContainsAny(compositionName, ".") {
//
//                                // Если имя ссылочного поля присутствует, то скопируем правила фильтрации для под запроса
//                                if strings.HasPrefix(compositionName, referenceFieldName) {
//                                    compositionName = strings.TrimPrefix(compositionName, referenceFieldName)
//
//                                    if compositionName == "" {
//                                        // Если после "." ни чего нет, то это означает фильтрацию текущего объекта
//                                        qopt.FilterCode = optVal
//
//                                    } else if !strings.ContainsAny(compositionName, ".") {
//                                        // Если точки нет - то проверяем Composition текущей сущности, иначе это относится к подчиненным объектам
//                                        // TODO - выделить общую часть из parseQueryOptionsReference
//                                        if qopt.Entity.GetComposition(compositionName) != nil {
//                                            // Сохраним код фильтрации Composition
//                                            if qopt.FilterCodeMap == nil {
//                                                qopt.FilterCodeMap = make(map[string]string)
//                                            }
//                                            qopt.FilterCodeMap[compositionName] = optVal
//                                        } else {
//                                            // Не существует такой Composition
//                                            return nil, nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - does not have Composition with name '%s', urlOption='%s=%s'", referenceEntity.Name, compositionName, optName, optVal)).PrintfError()
//                                        }
//                                        // TODO - выделить общую часть из parseQueryOptionsReference
//                                    } else {
//                                        // относится к подчиненным объектам - пробросим через рекурсию на уровень ниже
//                                        optFilterName := delimiterStartFilter + "." + compositionName + delimiterEnd // Сформируем правила для фильтрации
//                                        queryOptionsOut[optFilterName] = optVal
//                                    }
//                                }
//
//                            }
//                        }
//                    }
//                }
//            }
//        } // Сформируем правила фильтрации Composition
//
//        if popt, err = qopt.prepareProcessOptions(); err != nil {
//            return nil, nil, err
//        }
//
//        return qopt, popt, nil
//    }
//    return nil, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.storageMap != nil && referenceEntity != nil && queryOptionsIn != nil && referenceField != nil && referenceField.Reference() != nil {}", []interface{}{s, referenceEntity, queryOptionsIn, referenceField}).PrintfError()
//}
//
//// constructFieldsMap сформировать набор полей для ограничения возвращаемых полей
//func (s *Service) constructFieldsMap(requestID uint64, entity *_meta.Entity, fieldsName []string, nameFormat string, ignoreExtra bool, addRefFields bool) (fieldsMap _meta.FieldsMap, err error) {
//    if s != nil && entity != nil {
//        isRestrictFieldsTag := fieldsName != nil && len(fieldsName) > 0
//
//        if isRestrictFieldsTag {
//            var errDetail []string // ошибки агрегируем, чтобы потом отдать все сразу
//
//            fieldsMap = make(_meta.FieldsMap, len(entity.StructFields())) // максимальное количество равно всем полям структуры
//
//            for _, fieldName := range fieldsName {
//
//                // Ищем поля с учетом формата имени поля
//                if field := entity.FieldByTagName(nameFormat, fieldName); field != nil {
//
//                    fieldsMap[field.Name] = field
//
//                    // Если поле является ссылкой, то нужно добавить все поля этой ссылки, чтобы они выбирались из БД
//                    if field.InternalType == _meta.FIELD_TYPE_ASSOCIATION || field.InternalType == _meta.FIELD_TYPE_COMPOSITION {
//                        ref := field.Reference()
//                        if ref != nil {
//                            ref.AddFieldsToMap(&fieldsMap)
//                        } else {
//                            errDetail = append(errDetail, fmt.Sprintf("Entity '%s', Reference '%s' - empty 'Reference' pointer", entity.Name, field.Name))
//                        }
//                    }
//
//                    // Если поле входит в состав ссылки, то нужно добавить эту ссылку как поле и все поля этой ссылки
//                    if addRefFields {
//                        for _, ref := range field.ReferencesDef() {
//                            if ref != nil {
//                                ref.AddFieldsToMap(&fieldsMap)
//                            } else {
//                                errDetail = append(errDetail, fmt.Sprintf("Entity '%s', Reference '%s' - empty 'Reference' pointer", entity.Name, field.Name))
//                            }
//                        }
//                    }
//
//                } else {
//
//                    // не найденные поля игнорируем, если в них есть "." - это может относиться к вложенным сущностям
//                    if !(ignoreExtra || strings.ContainsAny(fieldName, ".")) {
//                        errDetail = append(errDetail, fmt.Sprintf("Entity '%s' has not have enabled field '%s', '%s'='%s'", entity.Name, fieldName, s.cfg.QueryOption.NameFormat, nameFormat))
//                    }
//                }
//            }
//
//            // Обработаем накопленные ошибки
//            if len(errDetail) > 0 {
//                myErr := _err.NewTyped(_err.ERR_ERROR, requestID, "Common error")
//                myErr.Details = errDetail
//                // вернем ожидаемый формат структуры
//                if rowExpected, err := s.newRowAll(requestID, entity, nil); err == nil {
//                    //myErr.Extra = rowExpected.PtrValue
//                    myErr.Extra = rowExpected.Value
//                }
//                return nil, myErr
//            }
//
//        } else {
//            // Если нет ограничения на состав полей, то передаем все
//            // TODO - копия снижает производительность. Проверить затраты
//            fieldsMap = entity.StructFieldsMap()
//        }
//
//        return fieldsMap, nil
//    }
//    return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil {}", []interface{}{s, entity}).PrintfError()
//}
