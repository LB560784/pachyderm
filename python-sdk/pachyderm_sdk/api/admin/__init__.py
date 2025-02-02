# Generated by the protocol buffer compiler.  DO NOT EDIT!
# sources: api/admin/admin.proto
# plugin: python-betterproto
# This file has been @generated
from dataclasses import dataclass
from typing import (
    TYPE_CHECKING,
    Dict,
    List,
    Optional,
)

import betterproto
import betterproto.lib.google.protobuf as betterproto_lib_google_protobuf
import grpc

from .. import (
    pfs as _pfs__,
    version as _version__,
)


if TYPE_CHECKING:
    import grpc


@dataclass(eq=False, repr=False)
class ClusterInfo(betterproto.Message):
    id: str = betterproto.string_field(1)
    deployment_id: str = betterproto.string_field(2)
    warnings_ok: bool = betterproto.bool_field(3)
    """True if the server is capable of generating warnings."""

    warnings: List[str] = betterproto.string_field(4)
    """Warnings about the client configuration."""

    proxy_host: str = betterproto.string_field(5)
    """The configured public URL of Pachyderm."""

    proxy_tls: bool = betterproto.bool_field(6)
    """True if Pachyderm is served over TLS (HTTPS)."""

    paused: bool = betterproto.bool_field(7)
    """True if this pachd is in "paused" mode."""

    web_resources: "WebResource" = betterproto.message_field(8)
    """Any HTTP links that the client might want to be aware of."""


@dataclass(eq=False, repr=False)
class InspectClusterRequest(betterproto.Message):
    client_version: "_version__.Version" = betterproto.message_field(1)
    """
    The version of the client that's connecting; used by the server to warn
    about too-old (or too-new!) clients.
    """

    current_project: "_pfs__.Project" = betterproto.message_field(2)
    """
    If CurrentProject is set, then InspectCluster will return an error if the
    project does not exist.
    """


@dataclass(eq=False, repr=False)
class WebResource(betterproto.Message):
    """WebResource contains URL prefixes of common HTTP functions."""

    archive_download_base_url: str = betterproto.string_field(1)
    """
    The base URL of the archive server; append a filename to this.  Empty if
    the archive server is not exposed.
    """

    create_pipeline_request_json_schema_url: str = betterproto.string_field(2)
    """
    Where to find the CreatePipelineRequest JSON schema; if this server is not
    accessible via a URL, then a link to Github is provided based on the baked-
    in version of the server.
    """


class ApiStub:
    def __init__(self, channel: "grpc.Channel"):
        self.__rpc_inspect_cluster = channel.unary_unary(
            "/admin_v2.API/InspectCluster",
            request_serializer=InspectClusterRequest.SerializeToString,
            response_deserializer=ClusterInfo.FromString,
        )

    def inspect_cluster(
        self,
        *,
        client_version: "_version__.Version" = None,
        current_project: "_pfs__.Project" = None
    ) -> "ClusterInfo":
        request = InspectClusterRequest()
        if client_version is not None:
            request.client_version = client_version
        if current_project is not None:
            request.current_project = current_project

        return self.__rpc_inspect_cluster(request)


class ApiBase:
    def inspect_cluster(
        self,
        client_version: "_version__.Version",
        current_project: "_pfs__.Project",
        context: "grpc.ServicerContext",
    ) -> "ClusterInfo":
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method not implemented!")
        raise NotImplementedError("Method not implemented!")

    __proto_path__ = "admin_v2.API"

    @property
    def __rpc_methods__(self):
        return {
            "InspectCluster": grpc.unary_unary_rpc_method_handler(
                self.inspect_cluster,
                request_deserializer=InspectClusterRequest.FromString,
                response_serializer=InspectClusterRequest.SerializeToString,
            ),
        }
