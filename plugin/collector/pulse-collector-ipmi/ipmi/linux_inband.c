#include <asm/ioctl.h>
#include <linux/ipmi.h>
#include <string.h>
#include <fcntl.h>
#include <errno.h>
#include <sys/ioctl.h>
#include <sys/time.h>

#include "linux_inband.h"

#define TIMEOUT 5


void IPMI_Syserr(struct IpmiStatusInfo *out) {
	out->system_error = errno;
	strerror_r(out->system_error, out->error_str, sizeof(out->error_str));
}


// error codes:
// 0 - ok
// <0 - invalid call
// >0 - errors from OS
int IPMI_BatchCommands(char *device, struct IpmiCommandInput *inputs,
	struct IpmiCommandOutput *outputs, int n, int n_sim,
	struct IpmiStatusInfo *info) {
		int fd, i, sent = 0, recvd = 0, readyFds;
		struct ipmi_ipmb_addr sendAddr={0}, recvAddr={0};
		struct ipmi_req request={0};
		struct ipmi_recv recv={0};
		struct timeval timeoutSend, timeoutRecv;
		fd_set fdset;
		char outData[1024];

		timeoutSend.tv_sec = TIMEOUT;
		timeoutSend.tv_usec = 0;

		timeoutRecv.tv_sec = TIMEOUT;
		timeoutRecv.tv_usec = 0;


		if (!info) {
			return -1;
		}

		for(i = 0; i < n; i++)
		{
			if (inputs[i].data_len < 2) {
				strcpy(info->error_str, "Supplied buffer too short in msg %d");
				return -2;
			}
		}

		fd = open(device, O_RDWR);
		if (fd < 0)
		{
			(info);
			return 100;
		}

		while(recvd < n) {
			if(sent < n && (sent-recvd) < n_sim) {

				sendAddr.addr_type = IPMI_IPMB_ADDR_TYPE;
				sendAddr.channel = inputs[sent].channel;
				sendAddr.slave_addr = inputs[sent].slave;
				sendAddr.lun = 0;

				request.addr = (char*)&sendAddr;
				request.addr_len = sizeof(sendAddr);

				request.msgid = sent;

				request.msg.netfn = inputs[sent].data[0];
				request.msg.cmd = inputs[sent].data[1];
				request.msg.data = &inputs[sent].data[2];
				request.msg.data_len = inputs[sent].data_len - 2;

				if (ioctl(fd, IPMICTL_SEND_COMMAND, &request) < 0) {
					IPMI_Syserr(info);
					close(fd);
					return 220;
				}

				sent++;
				continue;
			}
			//if we are at this point some messages are sent

			FD_ZERO(&fdset);
			FD_SET(fd, &fdset);

			if ( (readyFds = select(fd+1, &fdset, NULL, NULL, &timeoutRecv)) < 0) {
				IPMI_Syserr(info);
				close(fd);
				return 300;
			}

			if (readyFds < 1) {
				strcpy(info->error_str,"Timeout on read select.");
				close(fd);
				return 310;
			}

			recv.addr = (char*)&recvAddr;
			recv.addr_len = sizeof(recvAddr);

			recv.msg.data = outData;
			recv.msg.data_len = sizeof(outData);

			if (ioctl(fd, IPMICTL_RECEIVE_MSG_TRUNC, &recv) < 0) {
				IPMI_Syserr(info);
				close(fd);
				return 320;
			}

			// using memcpy here results in glibc dependency, so this simple for loop
			// avoids that
			for(i = 0; i < recv.msg.data_len; i++) {
				outputs[recv.msgid].data[i] = recv.msg.data[i];
			}
			outputs[recv.msgid].data_len = recv.msg.data_len;
			recvd++;

		}


		close(fd);
		return 0;
	}
