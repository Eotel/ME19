// ME19.cpp : コンソール アプリケーションのエントリ ポイントを定義します。
//

#include "stdafx.h"

#define	INTERVAL	100	
#define MAX_RETRY_COUNT	8	//QRコードがなくなってからQRコードをクリアするまでの回数
#define	FILENAME	"code.txt"
#define WRITE_RETRY_MAX	5	//書き込み再試行回数


#define CONSOLE_WINDOWTITLE "ME19Console"
#define CAPTURE_WINDOWTITLE "ME19Camera"

#define ME19QUITFILENAME "C:\\ME19\\ME19quit"

using namespace zbar;


//#ifdef _DEBUG
//	#define DISP_IMAGE
//	//#include <cv.h>
//	//#include <highgui.h>
//#else
//	#undef DISP_IMAGE
//#endif

//#define DISP_IMAGE


/* adapted from v4l2 spec */
#define FOURCC(a, b, c, d)                      \
    ((unsigned long)(a) | ((unsigned long)(b) << 8) |     \
     ((unsigned long)(c) << 16) | ((unsigned long)(d) << 24))

int writeCode(char* code);



BOOL WINAPI HandlerRoutine(DWORD dwCtrlType);
void CallbackSingnalControl(int sig);
bool FileExist(LPCTSTR lpFilename);


zbar_image_scanner_t*	scanner = NULL;
//zbar_video_t*	video;
CvCapture*	pCapture = NULL;

bool bLoopFlg = 0;

/*
 * コマンド引数
 * -delay: 読み取り無効カウント
 * -interval: 読み取り間隔(ms)
 */
int _tmain(int argc, _TCHAR* argv[])
{
//	int	ret;
	int	maxretrycount = MAX_RETRY_COUNT;
	int interval = INTERVAL;

	scanner = NULL;
	pCapture = NULL;

	//引数
	for(int i = 0; i < argc; i++){
		if(strcmp(argv[i], "-delay") == 0){
			i++;
			if(i < argc){
				maxretrycount = atoi(argv[i]);
			}
		}
		if(strcmp(argv[i], "-interval") == 0){
			i++;
			if(i < argc){
				interval = atoi(argv[i]);
			}
		}
	}


	// コンソールウィンドウの名前を変更してちょっと待つ
	SetConsoleTitle( CONSOLE_WINDOWTITLE );
	Sleep(50);


	//終了関数登録
	SetConsoleCtrlHandler(HandlerRoutine, TRUE);

	// Ctrl-C とかをキャッチする
	signal( SIGINT, CallbackSingnalControl);


    // initialize opencv camera
    pCapture = cvCaptureFromCAM( CV_CAP_ANY  );

	// initialize 
	scanner = zbar_image_scanner_create();
	if(scanner == NULL){
		printf("zbar_image_scanner_create error\n");
		return -1;
	}

	zbar_image_scanner_set_config(scanner, ZBAR_NONE, ZBAR_CFG_ENABLE, 1);

    // always check
    if ( !pCapture ) {
        fprintf( stderr, "Cannot open initialize webcam!\n" );
        return 1;
    }

#ifdef DISP_IMAGE

	// ウィンドウハンドル
	HWND hwParent = NULL;
	HWND hwCapture = NULL;
	long windowStyle = 0;

	cvNamedWindow( CAPTURE_WINDOWTITLE, 0 );

	hwCapture = (HWND) cvGetWindowHandle( CAPTURE_WINDOWTITLE );
	hwParent = GetParent(hwCapture);
//	ShowWindow( hwCapture, SW_MINIMIZE );	// 実はDebugウィンドウの中のウィンドウがminimizeするだけという…
	ShowWindow( hwParent, SW_MINIMIZE );


	//IplImage* showImage = cvCreateImage(cvSize(zbar_video_get_width(video), zbar_video_get_height(video)), IPL_DEPTH_8U, 3);
	//IplImage* showImage = cvQueryFrame( pCapture );
#endif

#define lastCodeBufLength 100

	bLoopFlg = TRUE;

	int retrycount = 0;
	char	lastCode[lastCodeBufLength];
	sprintf_s(lastCode, "");


	IplImage* cvCapImage = NULL;


	// zbar_image_t を作成
	zbar_image_t* zImage = NULL;
	zImage = zbar_image_create();
	assert(zImage);
	zbar_image_set_format( zImage, *(unsigned long*)"Y800");
	int width = 640;	//cvGrayImage->width;
	int height = 480;	//cvGrayImage->height;
	zbar_image_set_size( zImage, 640, 480);
	size_t imageSize = width * height;

	// zbar_image_t の中身（映像バッファ）を作成、割り当て
	unsigned char* pRaw = NULL;
	pRaw = (unsigned char*)malloc(imageSize);
	zbar_image_set_data( zImage, pRaw, imageSize, zbar_image_free_data);

	// メインループ
	while(bLoopFlg){
		Sleep(interval);

		// カメラから映像をキャプチャ
		if(pCapture){
			cvCapImage = cvQueryFrame( pCapture );
		}


		if(!cvCapImage){
			// イメージが取得出来なかった
			break;
		}
		else {
			// キャプチャしたカラーイメージをグレイスケールに変換
			IplImage* cvGrayImage = NULL;
			cvGrayImage = cvCreateImage( cvSize( cvCapImage->width, cvCapImage->height), 8, 1);
			cvCvtColor( cvCapImage, cvGrayImage, CV_BGR2GRAY);

			// zbar_image_t の画像バッファへコピー
			if(zImage){
				memcpy( pRaw, cvGrayImage->imageData, cvGrayImage->imageSize); 
			}

#ifdef DISP_IMAGE
			// カメラ映像をウィンドウに表示
			cvShowImage( CAPTURE_WINDOWTITLE, cvGrayImage);
#endif
			// グレイスケールイメージを破棄
			if( cvGrayImage ){
				cvReleaseImage( &cvGrayImage );
				cvGrayImage = NULL;
			}
		}

		if(zImage == NULL){
			//printf("video is not enabled or an error occurs\n");
			//break;
		}
		else {

			// ZBarのテンポラリイメージに… zImageを白黒にしてコピーしてる？
			zbar_image_t* zTmp = NULL;
			zTmp = zbar_image_convert( zImage, FOURCC('Y','8','0','0'));

			// ZBarでバーコード解析
			int n = 0;
			if(scanner){
				zbar_image_scanner_recycle_image( scanner, zImage);
				n = zbar_scan_image( scanner, zTmp);
			}

			if(n < 0){
				printf("scan error\n");
				break;
			}

			if(n > 0){
				//QRが読み取れた
				const zbar_symbol_t *sym = zbar_image_first_symbol(zTmp);
				const char* code = zbar_symbol_get_data(sym);
				if(strcmp(lastCode, "") == 0){
					//新たにQRが提示された
					printf("new code\n");
					printf("QRcode:%s\n", code);
					sprintf_s(lastCode, "%s", code);
					writeCode(lastCode);
					retrycount = 0;
				}
				else if(strcmp(lastCode, code) != 0){
					if(retrycount == maxretrycount){
						//違うQRが提示された
						printf("other code\n");
						printf("QRcode:%s\n", code);
						sprintf_s(lastCode, "%s", code);
						writeCode(lastCode);
						retrycount = 0;
					}else{
						retrycount++;
					}
				}else if(strcmp(lastCode, code) == 0){
					retrycount = 0;
				}
			}
			else if(n == 0){
				//読み取りなし
				if(strcmp(lastCode, "") != 0){
					if(retrycount == maxretrycount){
						//クリアする
						printf("remove code\n");
						sprintf_s(lastCode, "");
						writeCode(lastCode);
						retrycount = 0;
					}else{
						retrycount++;
					}
				}
			}

			// ZBarテンポラリイメージを破棄
			if(zTmp){
				zbar_image_destroy(zTmp);
				zTmp = NULL;
			}
		}

		// 何らかキー入力されたらループから抜ける
		if( cvWaitKey(1) >= 0 ){
			bLoopFlg = FALSE;
		}

		// C:のルートにファイルがあったらループから抜ける
		if( FileExist(ME19QUITFILENAME) ){
			// 即座にファイルを消す
			DeleteFile(ME19QUITFILENAME);
			bLoopFlg = FALSE;
		}

	}

#ifdef DISP_IMAGE
	// OpenCVのウィンドウを閉じる
	cvDestroyWindow( CAPTURE_WINDOWTITLE );
#endif

	// キャプチャデバイスを解放
	cvReleaseCapture( &pCapture );

	return 0;
}

int writeCode(char* code){
	FILE* fp;
	int	retry;
	bool success = false;

	for(retry = 0; retry < WRITE_RETRY_MAX; retry++){
		errno_t err = fopen_s( &fp, FILENAME, "w");

		if(fp){
			fwrite(code, sizeof(char), strlen(code), fp);
			fclose(fp);

			success = true;
			break;
		}

		Sleep(10);
	}

	if(success){
		return 0;
	}else{
		return 1;
	}
}

//
// 終了時処理
//
BOOL WINAPI HandlerRoutine(DWORD dwCtrlType){
	printf("exit\n");

	//int ret;

	////ビデオ停止
	//if(video){
	//	ret = zbar_video_enable(video, 0);
	//	if(ret < 0){
	//		printf("video disable error\n");
	//	}
	//}

	//return FALSE;

	bLoopFlg = FALSE;

	return TRUE;
}

// コンソールシグナル (Ctrl-C とか）を受け取る
void CallbackSingnalControl(int sig)
{
	if(sig==SIGINT){
		//再びシグナルを設定しない場合、次に発生した場合こなくなるっぽい
		//終わらせたい場合は、exit(0);とかここに書いちゃえばよい（後始末は忘れないように！）
		signal(SIGINT, CallbackSingnalControl);
		printf("catch signal %d\n", sig);
	}

	bLoopFlg = FALSE;
}

// ファイルの有無をチェック
bool FileExist(LPCTSTR lpFilename)
{
	bool ret;
	WIN32_FIND_DATA wfd;
	HANDLE hFile = FindFirstFile( lpFilename, &wfd);
	if(hFile==INVALID_HANDLE_VALUE)
		ret = false;	// 無い
	else
		ret = true;		// ハッケン
	FindClose(hFile);
	return ret;
}
