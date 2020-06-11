package denny

// standard grpc controllers have following signature
// func(s *ControllerPointer) Method(context, request) (response, error)
//func (s *HelloServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
//	var (
//		logger = log.New()
//	)
//
//	logger.Infof("request %s", logger.ToJsonString(in))
//	response := &pb.HelloResponse{
//		Reply: "hi",
//	}
//
//	return response, nil
//}
