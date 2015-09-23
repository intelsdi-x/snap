struct IpmiCommandInput {
	char data[1024];
	int data_len;
	short channel;
	unsigned char slave;
};

struct IpmiStatusInfo {
	char error_str[1024];
	int system_error;
};

struct IpmiCommandOutput {
	char data[1024];
	int data_len;
};

int IPMI_BatchCommands(char *device, struct IpmiCommandInput *inputs,
	struct IpmiCommandOutput *outputs, int n, int n_sim,
	struct IpmiStatusInfo *info);
