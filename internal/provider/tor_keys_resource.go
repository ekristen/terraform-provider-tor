package provider

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/crypto/sha3"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &TorKeys{}

func NewTor() resource.Resource {
	return &TorKeys{}
}

// TorKeys defines the data source implementation.
type TorKeys struct{}

// TorModel describes the data source data model.
type TorModel struct {
	ID         types.String `tfsdk:"id"`
	PrivateKey types.String `tfsdk:"private_key"`
	PublicKey  types.String `tfsdk:"public_key"`
	Address    types.String `tfsdk:"address"`
}

func (r *TorKeys) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keys"
}

func (r *TorKeys) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `tor` resource generates a new private/public key and onion address",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the resource",
				Computed:            true,
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "Private Key",
				Computed:            true,
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "Public Key",
				Computed:            true,
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "Onion Address",
				Computed:            true,
			},
		},
	}
}

func (r *TorKeys) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
}

func (r *TorKeys) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TorModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := uuid.NewUUID()
	if err != nil {
		resp.Diagnostics.AddError("failed to generate uuid", err.Error())
		return
	}

	publicKey, secretKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		resp.Diagnostics.AddError("failed to generate ed25519 key pair", err.Error())
		return
	}

	onionAddress := fmt.Sprintf("%s.onion", encodePublicKey(publicKey))

	// Set the computed value
	data.ID = types.StringValue(id.String())
	data.Address = types.StringValue(onionAddress)
	data.PublicKey = types.StringValue(base64.StdEncoding.EncodeToString(publicKey))
	data.PrivateKey = types.StringValue(base64.StdEncoding.EncodeToString(secretKey))

	fmt.Println(data.Address)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TorKeys) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TorModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TorKeys) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TorModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *TorKeys) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TorModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TorKeys) ImportState(_ context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError("import is not supported", "import is not supported")
}

const b32Lower = "abcdefghijklmnopqrstuvwxyz234567"

var b32Enc = base32.NewEncoding(b32Lower).WithPadding(base32.NoPadding)

func encodePublicKey(publicKey ed25519.PublicKey) string {
	// checksum = H(".onion checksum" || publicKey || version)
	var checksumBytes bytes.Buffer
	checksumBytes.Write([]byte(".onion checksum"))
	checksumBytes.Write(publicKey)
	checksumBytes.Write([]byte{0x03})
	checksum := sha3.Sum256(checksumBytes.Bytes())

	// onion_address = base32(publicKey || checksum || version)
	var onionAddressBytes bytes.Buffer
	onionAddressBytes.Write(publicKey)
	onionAddressBytes.Write(checksum[:2])
	onionAddressBytes.Write([]byte{0x03})
	onionAddress := b32Enc.EncodeToString(onionAddressBytes.Bytes())

	return onionAddress
}
