// ME19.cpp : �R���\�[�� �A�v���P�[�V�����̃G���g�� �|�C���g���`���܂��B
//

#include "stdafx.h"

#define	INTERVAL	100	
#define MAX_RETRY_COUNT	8	//QR�R�[�h���Ȃ��Ȃ��Ă���QR�R�[�h���N���A����܂ł̉�
#define	FILENAME	"code.txt"
#define WRITE_RETRY_MAX	5	//�������ݍĎ��s��


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
 * �R�}���h����
 * -delay: �ǂݎ�薳���J�E���g
 * -interval: �ǂݎ��Ԋu(ms)
 */
int _tmain(int argc, _TCHAR* argv[])
{
//	int	ret;
	int	maxretrycount = MAX_RETRY_COUNT;
	int interval = INTERVAL;

	scanner = NULL;
	pCapture = NULL;

	//����
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


	// �R���\�[���E�B���h�E�̖��O��ύX���Ă�����Ƒ҂�
	SetConsoleTitle( CONSOLE_WINDOWTITLE );
	Sleep(50);


	//�I���֐��o�^
	SetConsoleCtrlHandler(HandlerRoutine, TRUE);

	// Ctrl-C �Ƃ����L���b�`����
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

	// �E�B���h�E�n���h��
	HWND hwParent = NULL;
	HWND hwCapture = NULL;
	long windowStyle = 0;

	cvNamedWindow( CAPTURE_WINDOWTITLE, 0 );

	hwCapture = (HWND) cvGetWindowHandle( CAPTURE_WINDOWTITLE );
	hwParent = GetParent(hwCapture);
//	ShowWindow( hwCapture, SW_MINIMIZE );	// ����Debug�E�B���h�E�̒��̃E�B���h�E��minimize���邾���Ƃ����c
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


	// zbar_image_t ���쐬
	zbar_image_t* zImage = NULL;
	zImage = zbar_image_create();
	assert(zImage);
	zbar_image_set_format( zImage, *(unsigned long*)"Y800");
	int width = 640;	//cvGrayImage->width;
	int height = 480;	//cvGrayImage->height;
	zbar_image_set_size( zImage, 640, 480);
	size_t imageSize = width * height;

	// zbar_image_t �̒��g�i�f���o�b�t�@�j���쐬�A���蓖��
	unsigned char* pRaw = NULL;
	pRaw = (unsigned char*)malloc(imageSize);
	zbar_image_set_data( zImage, pRaw, imageSize, zbar_image_free_data);

	// ���C�����[�v
	while(bLoopFlg){
		Sleep(interval);

		// �J��������f�����L���v�`��
		if(pCapture){
			cvCapImage = cvQueryFrame( pCapture );
		}


		if(!cvCapImage){
			// �C���[�W���擾�o���Ȃ�����
			break;
		}
		else {
			// �L���v�`�������J���[�C���[�W���O���C�X�P�[���ɕϊ�
			IplImage* cvGrayImage = NULL;
			cvGrayImage = cvCreateImage( cvSize( cvCapImage->width, cvCapImage->height), 8, 1);
			cvCvtColor( cvCapImage, cvGrayImage, CV_BGR2GRAY);

			// zbar_image_t �̉摜�o�b�t�@�փR�s�[
			if(zImage){
				memcpy( pRaw, cvGrayImage->imageData, cvGrayImage->imageSize); 
			}

#ifdef DISP_IMAGE
			// �J�����f�����E�B���h�E�ɕ\��
			cvShowImage( CAPTURE_WINDOWTITLE, cvGrayImage);
#endif
			// �O���C�X�P�[���C���[�W��j��
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

			// ZBar�̃e���|�����C���[�W�Ɂc zImage�𔒍��ɂ��ăR�s�[���Ă�H
			zbar_image_t* zTmp = NULL;
			zTmp = zbar_image_convert( zImage, FOURCC('Y','8','0','0'));

			// ZBar�Ńo�[�R�[�h���
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
				//QR���ǂݎ�ꂽ
				const zbar_symbol_t *sym = zbar_image_first_symbol(zTmp);
				const char* code = zbar_symbol_get_data(sym);
				if(strcmp(lastCode, "") == 0){
					//�V����QR���񎦂��ꂽ
					printf("new code\n");
					printf("QRcode:%s\n", code);
					sprintf_s(lastCode, "%s", code);
					writeCode(lastCode);
					retrycount = 0;
				}
				else if(strcmp(lastCode, code) != 0){
					if(retrycount == maxretrycount){
						//�ႤQR���񎦂��ꂽ
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
				//�ǂݎ��Ȃ�
				if(strcmp(lastCode, "") != 0){
					if(retrycount == maxretrycount){
						//�N���A����
						printf("remove code\n");
						sprintf_s(lastCode, "");
						writeCode(lastCode);
						retrycount = 0;
					}else{
						retrycount++;
					}
				}
			}

			// ZBar�e���|�����C���[�W��j��
			if(zTmp){
				zbar_image_destroy(zTmp);
				zTmp = NULL;
			}
		}

		// ���炩�L�[���͂��ꂽ�烋�[�v���甲����
		if( cvWaitKey(1) >= 0 ){
			bLoopFlg = FALSE;
		}

		// C:�̃��[�g�Ƀt�@�C�����������烋�[�v���甲����
		if( FileExist(ME19QUITFILENAME) ){
			// �����Ƀt�@�C��������
			DeleteFile(ME19QUITFILENAME);
			bLoopFlg = FALSE;
		}

	}

#ifdef DISP_IMAGE
	// OpenCV�̃E�B���h�E�����
	cvDestroyWindow( CAPTURE_WINDOWTITLE );
#endif

	// �L���v�`���f�o�C�X�����
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
// �I��������
//
BOOL WINAPI HandlerRoutine(DWORD dwCtrlType){
	printf("exit\n");

	//int ret;

	////�r�f�I��~
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

// �R���\�[���V�O�i�� (Ctrl-C �Ƃ��j���󂯎��
void CallbackSingnalControl(int sig)
{
	if(sig==SIGINT){
		//�ĂуV�O�i����ݒ肵�Ȃ��ꍇ�A���ɔ��������ꍇ���Ȃ��Ȃ���ۂ�
		//�I��点�����ꍇ�́Aexit(0);�Ƃ������ɏ������Ⴆ�΂悢�i��n���͖Y��Ȃ��悤�ɁI�j
		signal(SIGINT, CallbackSingnalControl);
		printf("catch signal %d\n", sig);
	}

	bLoopFlg = FALSE;
}

// �t�@�C���̗L�����`�F�b�N
bool FileExist(LPCTSTR lpFilename)
{
	bool ret;
	WIN32_FIND_DATA wfd;
	HANDLE hFile = FindFirstFile( lpFilename, &wfd);
	if(hFile==INVALID_HANDLE_VALUE)
		ret = false;	// ����
	else
		ret = true;		// �n�b�P��
	FindClose(hFile);
	return ret;
}
