//
// C910CaptureTest.cpp
//
// Windows上でLogicool C910を使って1920x1080のキャプチャができることを確認。
// videoInput LibraryとOpenCV2.2を組み合わせて実装してます。
//
 
//
// OpenCV 2.2
// http://opencv.willowgarage.com/wiki/
//
#pragma warning(disable: 4819)
#pragma warning(disable: 4996)
#include <opencv/cv.h>
#include <opencv/highgui.h>
#ifdef _DEBUG
	#pragma comment(lib, "opencv_core220d.lib")
	#pragma comment(lib, "opencv_highgui220d.lib")
#else
	#pragma comment(lib, "opencv_core220.lib")
	#pragma comment(lib, "opencv_highgui220.lib")
#endif
 
//
// videoInput Library
// http://www.muonics.net/school/spring05/videoInput/
//
#include <videoInput.h>
#pragma comment(lib, "videoInput.lib")
#pragma comment(linker, "/NODEFAULTLIB:atlthunk")
#pragma comment(linker, "/NODEFAULTLIB:libcmt")
 
//
// videoInput Library wrapper class
//
class VideoInputLib
{
private:
	IplImage *img_;
	int camera_id_;
 
	videoInput VI;
 
public:
	typedef enum tagCameraSetting {
		BRIGHTNESS		= 0,
		CONTRAST		= 1,
		HUE			= 2,
		SATURATION		= 3,
		SHARPNESS		= 4,
		GAMMA			= 5,
		COLOR_ENABLE		= 6,
		WHITE_BALANCE		= 7,
		BACKLIGHT_COMPENSATION	= 8,
		GAIN			= 9,
 
		PAN			= 100,
		TILT			= 101,
		ROLL			= 102,
		ZOOM			= 103,
		EXPOSURE		= 104,
		IRIS			= 105,
		FOCUS			= 106
	} CameraSetting;
 
	typedef enum tagCameraSettingFlag {
		AUTO	= 1,
		MANUAL	= 2
	} CameraSettingFlag;
 
public:
	VideoInputLib() : img_(NULL), camera_id_(-1)
	{
		videoInput::setVerbose(false);
		VI.setUseCallback(true);
 
		// for debug...
		VI.listDevices();
	};
 
	virtual ~VideoInputLib()
	{
		stop();
	};
 
	bool start(const int &camera_id, const int &width, const int &height, const int &fps = 30)
	{
		stop();
 
		VI.setIdealFramerate(camera_id, fps);
		if (VI.setupDevice(camera_id, width, height, VI_COMPOSITE) == false) return false;
 
		this->camera_id_ = camera_id;
 
		img_ = cvCreateImage(cvSize(width, height), 8, 3);
		return true;
	};
 
	void stop()
	{
		if (this->camera_id_ != -1) {
			VI.stopDevice(this->camera_id_);
			this->camera_id_ = -1;
		}
		if (img_ != NULL) {
			cvReleaseImage(&img_);
			img_ = NULL;
		}
	};
 
	void grab()
	{
		//3番目、4番目の引数はRGBをBGRにするか、上下逆転させるかのフラグ
		VI.getPixels(camera_id_, (unsigned char*)(img_->imageData), false, true);  
	};
 
	IplImage* image()
	{
		return img_;
	}
 
	bool getSetting(long prop, long &min, long &max, long &stepping_delta, long &current_value, long &flags, long &default_value)
	{
		if (prop < 100) {
			return VI.getVideoSettingFilter(camera_id_, prop, min, max, stepping_delta, current_value, flags, default_value);
		}
		else {
			long camera_prop = prop - 100;
			return VI.getVideoSettingCamera(camera_id_, camera_prop, min, max, stepping_delta, current_value, flags, default_value);
		}
		return false;
	}
 
	bool setSetting(long prop, long val, long flags = NULL, bool use_default = false)
	{
		bool rv;
		long min, max, sd, cv, f, dv;
		if (prop < 100) {
			rv = VI.getVideoSettingFilter(camera_id_, prop, min, max, sd, cv, f, dv);
			if (rv == false) return false;
			return VI.setVideoSettingFilter(camera_id_, prop, (val/sd)*sd, flags, use_default);
		}
		else {
			long camera_prop = prop - 100;
			rv = VI.getVideoSettingCamera(camera_id_, camera_prop, min, max, sd, cv, f, dv);
			if (rv == false) return false;
			return VI.setVideoSettingCamera(camera_id_, camera_prop, (val/sd)*sd, flags, use_default);
		}
		return false;
	}
};
 
const char *window_name = "test";
 
//
// How to use
//
int main(int argc, char* argv[])
{
	VideoInputLib vi;
	IplImage *img;
	bool rv;
 
	// 初期化
	rv = vi.start(
		0,    // camera_id
		1920, // capture width
		1080  // capture height
		);
	if (rv== false) {
		fprintf(stderr, "vi.start() failed...\n");
		return false;
	}
 
	// キャプチャループ
	cvNamedWindow (window_name);
	while(1) {
		vi.grab();
		img = vi.image();
		cvShowImage(window_name, img);
 
		int c = cvWaitKey(20);
		if (c == 27) break;
	}
 
	// 終了処理
	vi.stop();
	cvDestroyAllWindows();
	
	return 0;
}