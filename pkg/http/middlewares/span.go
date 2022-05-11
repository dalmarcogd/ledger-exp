package middlewares

//func NewSpanHTTPMiddleware(spans tracing.Tracing, ignorePaths ...string) Middleware {
//	return func(handler http.Handler) http.Handler {
//		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
//			wrapWriter := &wrapResponseWriter{ResponseWriter: writer}
//			ctx := request.Context()
//			path := request.URL.Path
//			for _, ignorePath := range ignorePaths {
//				if strings.Contains(path, ignorePath) {
//					handler.ServeHTTP(wrapWriter, request)
//					return
//				}
//			}
//			obfuscatedURL := obfuscateURL(path)
//			name := fmt.Sprintf("%v %v", request.Method, obfuscatedURL)
//			options := []tracing.SpanOption{
//				tracer.Tag(tracing.HTTPMethod, request.Method),
//				tracer.Tag(tracing.HTTPURL, request.URL.String()),
//				tracer.Tag(tracing.ServiceName, spans.ServiceName()),
//				tracer.SpanType(tracing.SpanTypeWeb),
//				tracer.Tag(ext.ResourceName, name),
//			}
//			ctx, span, err := spans.Extract(ctx, tracer.HTTPHeadersCarrier(request.Header), options...)
//			if err != nil {
//				ctx, span = spans.Span(ctx, options...)
//			}
//			defer span.Finish()
//			span.SetOperationName("http.server")
//			request = request.WithContext(ctx)
//			handler.ServeHTTP(wrapWriter, request)
//			statusCode := wrapWriter.statusCode
//			span.SetTag(tracing.HTTPCode, strconv.Itoa(statusCode))
//			span.SetTag(tracing.HTTPResponseSize, strconv.Itoa(wrapWriter.bodySize))
//			if statusCode >= 400 {
//				span.SetTag(tracing.Error, true)
//			}
//		})
//	}
//}
//
//// SpanUnaryServerInterceptor returns a new unary server interceptor for span.
//func SpanUnaryServerInterceptor(
//	spans tracing.Tracing,
//	ignorePaths ...string,
//) grpc.UnaryServerInterceptor {
//	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
//		for _, path := range ignorePaths {
//			if path == info.FullMethod {
//				return handler(ctx, req)
//			}
//		}
//
//		options := []tracing.SpanOption{
//			tracer.Tag(tracing.ServiceName, spans.ServiceName()),
//			tracer.SpanType(tracing.SpanTypeRPC),
//			tracer.Tag(tracing.ResourceName, info.FullMethod),
//			tracer.Tag(tracing.GRPCMethodName, info.FullMethod),
//			tracer.Measured(),
//		}
//
//		md, _ := metadata.FromIncomingContext(ctx) // nil is ok
//		ctx, span, err := spans.Extract(ctx, mdCarrier(md), options...)
//		if err != nil {
//			ctx, span = spans.Span(ctx, options...)
//		}
//		defer span.Finish()
//		span.SetOperationName("grpc.server")
//		span.SetTag(tracing.GRPCMethodKind, "unary")
//
//		for k, v := range md {
//			span.SetTag(fmt.Sprintf("%s.%s", tracing.GRPCMetadataPefix, strings.ToLower(k)), v)
//		}
//		resp, err := handler(ctx, req)
//		if err == io.EOF || err == context.Canceled {
//			err = nil
//		}
//		errcode := status.Code(err)
//		if errcode == codes.OK {
//			err = nil
//		}
//		span.SetTag(tracing.GRPCCode, errcode.String())
//		finishOptions := []tracer.FinishOption{
//			tracer.WithError(err),
//		}
//		span.Finish(finishOptions...)
//		return resp, err
//	}
//}
//
//// mdCarrier implements tracer.TextMapWriter and tracer.TextMapReader on top
//// of gRPC's metadata, allowing it to be used as a span context carrier for
//// distributed tracing.
//type mdCarrier metadata.MD
//
//// Get will return the first entry in the metadata at the given key.
//func (mdc mdCarrier) Get(key string) string {
//	if m := mdc[key]; len(m) > 0 {
//		return m[0]
//	}
//	return ""
//}
//
//// Set will add the given value to the values found at key. Key will be lowercased to match
//// the metadata implementation.
//func (mdc mdCarrier) Set(key, val string) {
//	k := strings.ToLower(key) // as per google.golang.org/grpc/metadata/metadata.go
//	mdc[k] = append(mdc[k], val)
//}
//
//// ForeachKey will iterate over all key/value pairs in the metadata.
//func (mdc mdCarrier) ForeachKey(handler func(key, val string) error) error {
//	for k, vs := range mdc {
//		for _, v := range vs {
//			if err := handler(k, v); err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}
//
//func obfuscateURL(url string) string {
//	paths := strings.Split(url, "/")
//	for i, path := range paths {
//		if _, err := uuid.Parse(path); err == nil {
//			paths[i] = "{UUID}"
//			continue
//		}
//		if _, err := strconv.ParseInt(path, 10, 64); err == nil {
//			paths[i] = "{ID}"
//			continue
//		}
//	}
//	return strings.Join(paths, "/")
//}
